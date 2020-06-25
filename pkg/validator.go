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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
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
func NewValidator(db *sqlx.DB) *Validator {
	kvs := ipfsethdb.NewKeyValueStore(db)
	database := ipfsethdb.NewDatabase(db)
	return &Validator{
		kvs:           kvs,
		trieDB:        trie.NewDatabase(kvs),
		stateDatabase: state.NewDatabase(database),
	}
}

// ValidateTrie returns whether or not the trie for the provided root hash is valid and complete
// Validating the completeness of a modified merkle patricia trie requires traversing the entire trie and verifying that
// every node is present, this is an expensive operation
func (v *Validator) ValidateTrie(root common.Hash) (bool, error) {
	// Generate the state.NodeIterator for this root
	snapshotTree := snapshot.New(v.kvs, v.trieDB, 0, root, false)
	stateDB, err := state.New(common.Hash{}, v.stateDatabase, snapshotTree)
	if err != nil {
		return false, err
	}
	it := state.NewNodeIterator(stateDB)
	for it.Next() {
		// iterate through entire trie
		// it.Next() will return false when we have either completed iteration of the entire trie or have ran into an error
		// if we are able to iterate through the entire trie without error then the trie is complete
	}
	if it.Error != nil {
		return false, it.Error
	}
	return true, nil
}
