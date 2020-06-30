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
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/jmoiron/sqlx"

	"github.com/vulcanize/pg-ipfs-ethdb"
)

// Validator is used for validating Ethereum state and storage tries on PG-IPFS
type Validator struct {
	kvs           ethdb.KeyValueStore
	trieDB        *trie.Database
	stateDatabase state.Database
}

// NewValidator returns a new trie validator
// Validating the completeness of a modified merkle patricia tries requires traversing the entire trie and verifying that
// every node is present, this is an expensive operation
func NewValidator(db *sqlx.DB) *Validator {
	kvs := ipfsethdb.NewKeyValueStore(db)
	database := ipfsethdb.NewDatabase(db)
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
	stateDB, err := state.New(common.Hash{}, v.stateDatabase, nil)
	if err != nil {
		return err
	}
	it := state.NewNodeIterator(stateDB)
	// state.NodeIterator won't throw an error if we can't find the root node
	// check if it exists first
	exists, err := v.kvs.Has(stateRoot.Bytes())
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("root node for hash %s does not exist in database", stateRoot.Hex())
	}
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
