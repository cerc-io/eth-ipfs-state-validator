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
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ipfs/go-blockservice"
	"github.com/jmoiron/sqlx"
	"github.com/mailgun/groupcache/v2"

	ipfsethdb "github.com/vulcanize/ipfs-ethdb"
	pgipfsethdb "github.com/vulcanize/ipfs-ethdb/postgres"
)

// Validator is used for validating Ethereum state and storage tries on PG-IPFS
type Validator struct {
	kvs           ethdb.KeyValueStore
	trieDB        *trie.Database
	stateDatabase state.Database
	db            *pgipfsethdb.Database
}

// NewPGIPFSValidator returns a new trie validator ontop of a connection pool for an IPFS backing Postgres database
func NewPGIPFSValidator(db *sqlx.DB) *Validator {
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

	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(kvs),
		stateDatabase: state.NewDatabase(database),
		db:            database.(*pgipfsethdb.Database),
	}
}

func (v *Validator) GetCacheStats() groupcache.Stats {
	return v.db.GetCacheStats()
}

// NewIPFSValidator returns a new trie validator ontop of an IPFS blockservice
func NewIPFSValidator(bs blockservice.BlockService) *Validator {
	kvs := ipfsethdb.NewKeyValueStore(bs)
	database := ipfsethdb.NewDatabase(bs)
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(kvs),
		stateDatabase: state.NewDatabase(database),
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

// ValidateTrie returns an error if the state and storage tries for the provided state root cannot be confirmed as complete
// This does consider child storage tries
func (v *Validator) ValidateTrie(stateRoot common.Hash) error {
	// Generate the state.NodeIterator for this root
	stateDB, err := state.New(stateRoot, v.stateDatabase, nil)
	if err != nil {
		return err
	}
	it := state.NewNodeIterator(stateDB)
	for it.Next() {
		// iterate through entire state trie and descendent storage tries
		// it.Next() will return false when we have either completed iteration of the entire trie or have ran into an error (e.g. a missing node)
		// if we are able to iterate through the entire trie without error then the trie is complete
	}
	return it.Error
}

// ValidateStateTrie returns an error if the state trie for the provided state root cannot be confirmed as complete
// This does not consider child storage tries
func (v *Validator) ValidateStateTrie(stateRoot common.Hash) error {
	// Generate the trie.NodeIterator for this root
	t, err := v.stateDatabase.OpenTrie(stateRoot)
	if err != nil {
		return err
	}
	it := t.NodeIterator(nil)
	for it.Next(true) {
		// iterate through entire state trie
		// it.Next() will return false when we have either completed iteration of the entire trie or have ran into an error (e.g. a missing node)
		// if we are able to iterate through the entire trie without error then the trie is complete
	}
	return it.Error()
}

// ValidateStorageTrie returns an error if the storage trie for the provided storage root and contract address cannot be confirmed as complete
func (v *Validator) ValidateStorageTrie(address common.Address, storageRoot common.Hash) error {
	// Generate the state.NodeIterator for this root
	addrHash := crypto.Keccak256Hash(address.Bytes())
	t, err := v.stateDatabase.OpenStorageTrie(addrHash, storageRoot)
	if err != nil {
		return err
	}
	it := t.NodeIterator(nil)
	for it.Next(true) {
		// iterate through entire storage trie
		// it.Next() will return false when we have either completed iteration of the entire trie or have ran into an error (e.g. a missing node)
		// if we are able to iterate through the entire trie without error then the trie is complete
	}
	return it.Error()
}

// Close implements io.Closer
// it deregisters the groupcache name
func (v *Validator) Close() error {
	groupcache.DeregisterGroup("kv")
	groupcache.DeregisterGroup("db")
	return nil
}
