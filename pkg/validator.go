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
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"

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
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	ipfsethdb "github.com/cerc-io/ipfs-ethdb/v4"
	pgipfsethdb "github.com/cerc-io/ipfs-ethdb/v4/postgres"
	nodeiter "github.com/ethereum/go-ethereum/trie/concurrent_iterator"
	"github.com/ethereum/go-ethereum/trie/concurrent_iterator/tracker"
)

// Validator is used for validating Ethereum state and storage tries on PG-IPFS
type Validator struct {
	kvs           ethdb.KeyValueStore
	trieDB        *trie.Database
	stateDatabase state.Database
	db            *pgipfsethdb.Database

	params Params
}

type Params struct {
	Workers        uint
	RecoveryFormat string // %s substituted with traversal type
}

var (
	DefaultRecoveryFormat = "./recover_validate_%s"
	emptyCodeHash         = crypto.Keccak256(nil)
)

type KVSWithAncient struct {
	kvs ethdb.KeyValueStore
	ethdb.Database
}

func NewKVSDatabaseWithAncient(kvs ethdb.KeyValueStore) ethdb.Database {
	return &KVSWithAncient{
		kvs: kvs,
	}
}

// NewPGIPFSValidator returns a new trie validator ontop of a connection pool for an IPFS backing Postgres database
func NewPGIPFSValidator(db *sqlx.DB, par Params) *Validator {
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

	normalizeParams(&par)
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(NewKVSDatabaseWithAncient(kvs)),
		stateDatabase: state.NewDatabase(database),
		db:            database.(*pgipfsethdb.Database),
		params:        par,
	}
}

func (v *Validator) GetCacheStats() groupcache.Stats {
	return v.db.GetCacheStats()
}

// NewIPFSValidator returns a new trie validator ontop of an IPFS blockservice
func NewIPFSValidator(bs blockservice.BlockService, par Params) *Validator {
	kvs := ipfsethdb.NewKeyValueStore(bs)
	database := ipfsethdb.NewDatabase(bs)
	normalizeParams(&par)
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(NewKVSDatabaseWithAncient(kvs)),
		stateDatabase: state.NewDatabase(database),
		params:        par,
	}
}

// NewValidator returns a new trie validator
// Validating the completeness of a modified merkle patricia tries requires traversing the entire trie and verifying that
// every node is present, this is an expensive operation
func NewValidator(kvs ethdb.KeyValueStore, database ethdb.Database) *Validator {
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(NewKVSDatabaseWithAncient(kvs)),
		stateDatabase: state.NewDatabase(database),
	}
}

// Ensure params are valid
func normalizeParams(p *Params) {
	if p.Workers == 0 {
		p.Workers = 1
	}
	if len(p.RecoveryFormat) == 0 {
		p.RecoveryFormat = DefaultRecoveryFormat
	}
}

// ValidateTrie returns an error if the state and storage tries for the provided state root cannot be confirmed as complete
// This does consider child storage tries
func (v *Validator) ValidateTrie(stateRoot common.Hash) error {
	t, err := v.stateDatabase.OpenTrie(stateRoot)
	if err != nil {
		return err
	}
	iterate := func(it trie.NodeIterator) error { return v.iterate(it, true) }
	return iterateTracked(t, fmt.Sprintf(v.params.RecoveryFormat, fullTraversal), v.params.Workers, iterate)
}

// ValidateStateTrie returns an error if the state trie for the provided state root cannot be confirmed as complete
// This does not consider child storage tries
func (v *Validator) ValidateStateTrie(stateRoot common.Hash) error {
	// Generate the trie.NodeIterator for this root
	t, err := v.stateDatabase.OpenTrie(stateRoot)
	if err != nil {
		return err
	}
	iterate := func(it trie.NodeIterator) error { return v.iterate(it, false) }
	return iterateTracked(t, fmt.Sprintf(v.params.RecoveryFormat, stateTraversal), v.params.Workers, iterate)
}

// ValidateStorageTrie returns an error if the storage trie for the provided storage root and contract address cannot be confirmed as complete
func (v *Validator) ValidateStorageTrie(stateRoot common.Hash, address common.Address, storageRoot common.Hash) error {
	// Generate the state.NodeIterator for this root
	addrHash := crypto.Keccak256Hash(address.Bytes())
	t, err := v.stateDatabase.OpenStorageTrie(stateRoot, addrHash, storageRoot)
	if err != nil {
		return err
	}
	iterate := func(it trie.NodeIterator) error { return v.iterate(it, false) }
	return iterateTracked(t, fmt.Sprintf(v.params.RecoveryFormat, storageTraversal), v.params.Workers, iterate)
}

// Close implements io.Closer
// it deregisters the groupcache name
func (v *Validator) Close() error {
	groupcache.DeregisterGroup("kv")
	groupcache.DeregisterGroup("db")
	return nil
}

// Traverses one iterator fully
// If storage = true, also traverse storage tries for each leaf.
func (v *Validator) iterate(it trie.NodeIterator, storage bool) error {
	// Iterate through entire state trie. it.Next() will return false when we have
	// either completed iteration of the entire trie or run into an error (e.g. a
	// missing node). If we are able to iterate through the entire trie without error
	// then the trie is complete.
	for it.Next(true) {
		// This block adapted from geth - core/state/iterator.go
		// If storage is not requested, or the state trie node is an internal entry, skip
		if !storage || !it.Leaf() {
			continue
		}
		// Otherwise we've reached an account node, initiate data iteration
		var account types.StateAccount
		if err := rlp.Decode(bytes.NewReader(it.LeafBlob()), &account); err != nil {
			return err
		}
		dataTrie, err := v.stateDatabase.OpenStorageTrie(common.HexToHash(viper.GetString("validator.stateRoot")), common.BytesToHash(it.LeafKey()), account.Root)
		if err != nil {
			return err
		}
		dataIt := dataTrie.NodeIterator(nil)
		if !bytes.Equal(account.CodeHash, emptyCodeHash) {
			addrHash := common.BytesToHash(it.LeafKey())
			_, err := v.stateDatabase.ContractCode(addrHash, common.BytesToHash(account.CodeHash))
			if err != nil {
				return fmt.Errorf("code %x: %w (path %x)", account.CodeHash, err, nodeiter.HexToKeyBytes(it.Path()))
			}
		}
		for dataIt.Next(true) {
		}
		if dataIt.Error() != nil {
			return fmt.Errorf("data iterator error (path %x): %w", nodeiter.HexToKeyBytes(dataIt.Path()), dataIt.Error())
		}
	}
	return it.Error()
}

// Traverses each iterator in a separate goroutine.
// Dumps to a recovery file on failure or interrupt.
func iterateTracked(tree state.Trie, recoveryFile string, iterCount uint, fn func(trie.NodeIterator) error) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	tracker := tracker.New(recoveryFile, iterCount)
	tracker.CaptureSignal(cancelCtx)
	halt := func() {
		if err := tracker.HaltAndDump(); err != nil {
			log.Errorf("failed to write recovery file: %v", err)
		}
	}

	// attempt to restore from recovery file if it exists
	iters, err := tracker.Restore(tree)
	if err != nil {
		return err
	}
	if iterCount < uint(len(iters)) {
		return fmt.Errorf("recovered too many iterators: got %d, expected %d", len(iters), iterCount)
	}

	if iters == nil { // nothing restored
		iters = nodeiter.SubtrieIterators(tree, iterCount)
		for i, it := range iters {
			iters[i] = tracker.Tracked(it, nil)
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	defer halt()

	for _, it := range iters {
		func(it trie.NodeIterator) {
			g.Go(func() error { return fn(it) })
		}(it)
	}
	return g.Wait()
}
