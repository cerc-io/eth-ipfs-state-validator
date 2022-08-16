// VulcanizeDB
// Copyright © 2020 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package validator

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ipfs/go-blockservice"
	"github.com/jmoiron/sqlx"
	"github.com/mailgun/groupcache/v2"

	nodeiter "github.com/vulcanize/go-eth-state-node-iterator"
	ipfsethdb "github.com/vulcanize/ipfs-ethdb/v4"
	pgipfsethdb "github.com/vulcanize/ipfs-ethdb/v4/postgres"
)

// Validator is used for validating Ethereum state and storage tries on PG-IPFS
type Validator struct {
	kvs           ethdb.KeyValueStore
	trieDB        *trie.Database
	stateDatabase state.Database
	db            *pgipfsethdb.Database

	iterWorkers uint
}

var emptyCodeHash = crypto.Keccak256(nil)

// NewPGIPFSValidator returns a new trie validator ontop of a connection pool for an IPFS backing Postgres database
func NewPGIPFSValidator(db *sqlx.DB, workers uint) *Validator {
	kvs := pgipfsethdb.NewKeyValueStore(db, pgipfsethdb.CacheConfig{
		Name:           "kv",
		Size:           16 * 1000 * 1000, // 16MB
		ExpiryDuration: time.Hour * 8,    // 8 hours
	})

	database := pgipfsethdb.NewDatabase(db, pgipfsethdb.CacheConfig{
		Name:           "db",
		Size:           16 * 1000 * 1000, // 16MB
		ExpiryDuration: time.Hour * 8,    // 8 hours
	})

	if workers == 0 {
		workers = 1
	}
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(kvs),
		stateDatabase: state.NewDatabase(database),
		db:            database.(*pgipfsethdb.Database),
		iterWorkers:   workers,
	}
}

func (v *Validator) GetCacheStats() groupcache.Stats {
	return v.db.GetCacheStats()
}

// NewIPFSValidator returns a new trie validator ontop of an IPFS blockservice
func NewIPFSValidator(bs blockservice.BlockService, workers uint) *Validator {
	kvs := ipfsethdb.NewKeyValueStore(bs)
	database := ipfsethdb.NewDatabase(bs)
	if workers == 0 {
		workers = 1
	}
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(kvs),
		stateDatabase: state.NewDatabase(database),
		iterWorkers:   workers,
	}
}

// NewValidator returns a new trie validator
// Validating the completeness of a modified merkle patricia tries requires traversing the entire trie and verifying that
// every node is present, this is an expensive operation
func NewValidator(kvs ethdb.KeyValueStore, database ethdb.Database) *Validator {
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(kvs),
		stateDatabase: state.NewDatabase(database),
	}
}

// Traverses each iterator in a separate goroutine.
// If storage = true, also traverse storage tries for each leaf.
func (v *Validator) iterateAsync(iters []trie.NodeIterator, storage bool) error {
	var wg sync.WaitGroup
	errors := make(chan error)
	for _, it := range iters {
		wg.Add(1)
		go func(it trie.NodeIterator) {
			defer wg.Done()
			for it.Next(true) {
				// Iterate through entire state trie. it.Next() will return false when we have
				// either completed iteration of the entire trie or run into an error (e.g. a
				// missing node). If we are able to iterate through the entire trie without error
				// then the trie is complete.

				// If storage is not requested, or the state trie node is an internal entry, leave as is
				if !storage || !it.Leaf() {
					continue
				}
				// Otherwise we've reached an account node, initiate data iteration
				var account types.StateAccount
				if err := rlp.Decode(bytes.NewReader(it.LeafBlob()), &account); err != nil {
					errors <- err
					break
				}
				dataTrie, err := v.stateDatabase.OpenStorageTrie(common.BytesToHash(it.LeafKey()), account.Root)
				if err != nil {
					errors <- err
					break
				}
				dataIt := dataTrie.NodeIterator(nil)
				if !bytes.Equal(account.CodeHash, emptyCodeHash) {
					addrHash := common.BytesToHash(it.LeafKey())
					_, err := v.stateDatabase.ContractCode(addrHash, common.BytesToHash(account.CodeHash))
					if err != nil {
						errors <- fmt.Errorf("code %x: %v", account.CodeHash, err)
						break
					}
				}
				for dataIt.Next(true) {
				}
				if err = dataIt.Error(); err != nil {
					errors <- err
					break
				}

			}
			if it.Error() != nil {
				errors <- it.Error()
			}
		}(it)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	var err error
	select {
	case err = <-errors:
	case <-done:
		close(errors)
	}
	return err
}

// ValidateTrie returns an error if the state and storage tries for the provided state root cannot be confirmed as complete
// This does consider child storage tries
func (v *Validator) ValidateTrie(stateRoot common.Hash) error {
	t, err := v.stateDatabase.OpenTrie(stateRoot)
	if err != nil {
		return err
	}
	iters := nodeiter.SubtrieIterators(t, v.iterWorkers)
	return v.iterateAsync(iters, true)
}

// ValidateStateTrie returns an error if the state trie for the provided state root cannot be confirmed as complete
// This does not consider child storage tries
func (v *Validator) ValidateStateTrie(stateRoot common.Hash) error {
	// Generate the trie.NodeIterator for this root
	t, err := v.stateDatabase.OpenTrie(stateRoot)
	if err != nil {
		return err
	}
	iters := nodeiter.SubtrieIterators(t, v.iterWorkers)
	return v.iterateAsync(iters, false)
}

// ValidateStorageTrie returns an error if the storage trie for the provided storage root and contract address cannot be confirmed as complete
func (v *Validator) ValidateStorageTrie(address common.Address, storageRoot common.Hash) error {
	// Generate the state.NodeIterator for this root
	addrHash := crypto.Keccak256Hash(address.Bytes())
	t, err := v.stateDatabase.OpenStorageTrie(addrHash, storageRoot)
	if err != nil {
		return err
	}
	iters := nodeiter.SubtrieIterators(t, v.iterWorkers)
	return v.iterateAsync(iters, false)
}

// Close implements io.Closer
// it deregisters the groupcache name
func (v *Validator) Close() error {
	groupcache.DeregisterGroup("kv")
	groupcache.DeregisterGroup("db")
	return nil
}
