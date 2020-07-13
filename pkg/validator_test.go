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

package validator_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ipfs/go-cid/_rsrch/cidiface"
	"github.com/jmoiron/sqlx"
	"github.com/multiformats/go-multihash"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/eth-ipfs-state-validator/pkg"
	"github.com/vulcanize/ipfs-ethdb/postgres"
)

var (
	contractAddr      = common.HexToAddress("0xaE9BEa628c4Ce503DcFD7E305CaB4e29E7476592")
	slot0StorageValue = common.Hex2Bytes("94703c4b2bd70c169f5717101caee543299fc946c7")
	slot1StorageValue = common.Hex2Bytes("01")
	nullCodeHash      = crypto.Keccak256Hash([]byte{})
	emptyRootNode, _  = rlp.EncodeToBytes([]byte{})
	emptyContractRoot = crypto.Keccak256Hash(emptyRootNode)

	stateBranchRootNode, _ = rlp.EncodeToBytes([]interface{}{
		crypto.Keccak256(bankAccountLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(minerAccountLeafNode),
		crypto.Keccak256(contractAccountLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(account2LeafNode),
		[]byte{},
		crypto.Keccak256(account1LeafNode),
		[]byte{},
		[]byte{},
	})
	stateRoot = crypto.Keccak256Hash(stateBranchRootNode)

	contractAccount, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0xaaea5efba4fd7b45d7ec03918ac5d8b31aa93b48986af0e6b591f0f087c80127").Bytes(),
		Root:     crypto.Keccak256Hash(storageBranchRootNode),
	})
	contractAccountLeafNode, _ = rlp.EncodeToBytes([]interface{}{
		common.Hex2Bytes("3114658a74d9cc9f7acf2c5cd696c3494d7c344d78bfec3add0d91ec4e8d1c45"),
		contractAccount,
	})

	minerAccount, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    0,
		Balance:  big.NewInt(1000),
		CodeHash: nullCodeHash.Bytes(),
		Root:     emptyContractRoot,
	})
	minerAccountLeafNode, _ = rlp.EncodeToBytes([]interface{}{
		common.Hex2Bytes("3380c7b7ae81a58eb98d9c78de4a1fd7fd9535fc953ed2be602daaa41767312a"),
		minerAccount,
	})

	account1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    2,
		Balance:  big.NewInt(1000),
		CodeHash: nullCodeHash.Bytes(),
		Root:     emptyContractRoot,
	})
	account1LeafNode, _ = rlp.EncodeToBytes([]interface{}{
		common.Hex2Bytes("3926db69aaced518e9b9f0f434a473e7174109c943548bb8f23be41ca76d9ad2"),
		account1,
	})

	account2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    0,
		Balance:  big.NewInt(1000),
		CodeHash: nullCodeHash.Bytes(),
		Root:     emptyContractRoot,
	})
	account2LeafNode, _ = rlp.EncodeToBytes([]interface{}{
		common.Hex2Bytes("3957f3e2f04a0764c3a0491b175f69926da61efbcc8f61fa1455fd2d2b4cdd45"),
		account2,
	})

	bankAccount, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    2,
		Balance:  big.NewInt(1000),
		CodeHash: nullCodeHash.Bytes(),
		Root:     emptyContractRoot,
	})
	bankAccountLeafNode, _ = rlp.EncodeToBytes([]interface{}{
		common.Hex2Bytes("30bf49f440a1cd0527e4d06e2765654c0f56452257516d793a9b8d604dcfdf2a"),
		bankAccount,
	})

	storageBranchRootNode, _ = rlp.EncodeToBytes([]interface{}{
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot0StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		crypto.Keccak256(slot1StorageLeafNode),
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
		[]byte{},
	})
	storageRoot = crypto.Keccak256Hash(storageBranchRootNode)

	slot0StorageLeafNode, _ = rlp.EncodeToBytes([]interface{}{
		common.Hex2Bytes("390decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563"),
		slot0StorageValue,
	})
	slot1StorageLeafNode, _ = rlp.EncodeToBytes([]interface{}{
		common.Hex2Bytes("310e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6"),
		slot1StorageValue,
	})

	trieStateNodes = [][]byte{
		stateBranchRootNode,
		bankAccountLeafNode,
		minerAccountLeafNode,
		contractAccountLeafNode,
		account1LeafNode,
		account2LeafNode,
	}
	trieStorageNodes = [][]byte{
		storageBranchRootNode,
		slot0StorageLeafNode,
		slot1StorageLeafNode,
	}

	missingRootStateNodes = [][]byte{
		bankAccountLeafNode,
		minerAccountLeafNode,
		contractAccountLeafNode,
		account1LeafNode,
		account2LeafNode,
	}
	missingRootStorageNodes = [][]byte{
		slot0StorageLeafNode,
		slot1StorageLeafNode,
	}

	missingNodeStateNodes = [][]byte{
		stateBranchRootNode,
		bankAccountLeafNode,
		minerAccountLeafNode,
		contractAccountLeafNode,
		account2LeafNode,
	}
	missingNodeStorageNodes = [][]byte{
		storageBranchRootNode,
		slot1StorageLeafNode,
	}
)

var (
	v   *validator.Validator
	db  *sqlx.DB
	err error
)

var _ = Describe("PG-IPFS Validator", func() {
	BeforeEach(func() {
		db, err = pgipfsethdb.TestDB()
		Expect(err).ToNot(HaveOccurred())
		v = validator.NewPGIPFSValidator(db)
	})
	Describe("ValidateTrie", func() {
		AfterEach(func() {
			err = validator.ResetTestDB(db)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Returns an error the state root node is missing", func() {
			loadTrie(missingRootStateNodes, trieStorageNodes)
			err = v.ValidateTrie(stateRoot)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist in database"))
		})
		It("Fails to return an error if the storage root node is missing", func() {
			// NOTE this failure was not expected and renders this approach unreliable, this is an issue with the go-ethereum core/state/iterator.NodeIterator
			loadTrie(trieStateNodes, missingRootStorageNodes)
			err = v.ValidateTrie(stateRoot)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Fails to return an error if the entire state (state trie and storage tries) cannot be validated", func() {
			// NOTE this failure was not expected and renders this approach unreliable, this is an issue with the go-ethereum core/state/iterator.NodeIterator
			loadTrie(missingNodeStateNodes, trieStorageNodes)
			err = v.ValidateTrie(stateRoot)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Fails to return an error if the entire state (state trie and storage tries) cannot be validated", func() {
			// NOTE this failure was not expected and renders this approach unreliable, this is an issue with the go-ethereum core/state/iterator.NodeIterator
			loadTrie(trieStateNodes, missingNodeStorageNodes)
			err = v.ValidateTrie(stateRoot)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Returns no error if the entire state (state trie and storage tries) can be validated", func() {
			loadTrie(trieStateNodes, trieStorageNodes)
			err = v.ValidateTrie(stateRoot)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ValidateStateTrie", func() {
		AfterEach(func() {
			err = validator.ResetTestDB(db)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Returns an error the state root node is missing", func() {
			loadTrie(missingRootStateNodes, nil)
			err = v.ValidateStateTrie(stateRoot)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing trie node"))
		})
		It("Returns an error if the entire state trie cannot be validated", func() {
			loadTrie(missingNodeStateNodes, nil)
			err = v.ValidateStateTrie(stateRoot)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing trie node"))
		})
		It("Returns no error if the entire state trie can be validated", func() {
			loadTrie(trieStateNodes, nil)
			err = v.ValidateStateTrie(stateRoot)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ValidateStorageTrie", func() {
		AfterEach(func() {
			err = validator.ResetTestDB(db)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Returns an error the storage root node is missing", func() {
			loadTrie(nil, missingRootStorageNodes)
			err = v.ValidateStorageTrie(contractAddr, storageRoot)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing trie node"))
		})
		It("Returns an error if the entire storage trie cannot be validated", func() {
			loadTrie(nil, missingNodeStorageNodes)
			err = v.ValidateStorageTrie(contractAddr, storageRoot)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing trie node"))
		})
		It("Returns no error if the entire storage trie can be validated", func() {
			loadTrie(nil, trieStorageNodes)
			err = v.ValidateStorageTrie(contractAddr, storageRoot)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func loadTrie(stateNodes, storageNodes [][]byte) {
	tx, err := db.Beginx()
	Expect(err).ToNot(HaveOccurred())
	for _, node := range stateNodes {
		_, err := validator.PublishRaw(tx, cid.EthStateTrie, multihash.KECCAK_256, node)
		Expect(err).ToNot(HaveOccurred())
	}
	for _, node := range storageNodes {
		_, err := validator.PublishRaw(tx, cid.EthStorageTrie, multihash.KECCAK_256, node)
		Expect(err).ToNot(HaveOccurred())
	}
	err = tx.Commit()
	Expect(err).ToNot(HaveOccurred())
}
