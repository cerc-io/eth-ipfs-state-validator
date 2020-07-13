// Copyright © 2020 Vulcanize, Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	_ "github.com/lib/pq" //postgres driver
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vulcanize/eth-ipfs-state-validator/pkg"
)

// validateTrieCmd represents the validateTrie command
var validateTrieCmd = &cobra.Command{
	Use:   "validateTrie",
	Short: "Validate completeness of state data on IPFS",
	Long: `This command is used to validate the completeness of state data corresponding specific to a specific root

If an ipfs-path is provided it will use a blockservice, otherwise it expects Postgres db configuration in a linked config file.

It can operate at three levels: 

"full" validates completeness of the entire state corresponding to a provided state root, including both state and storage tries

./eth-ipfs-state-validator validateTrie --config={path to db config} --type=full --state-root={state root hex string}


"state" validates completeness of the state trie corresponding to a provided state root, excluding the storage tries

./eth-ipfs-state-validator validateTrie --config={path to db config} --type=state --state-root={state root hex string}


"storage" validates completeness of only the storage trie corresponding to a provided storage root and contract address

./eth-ipfs-state-validator validateTrie --config={path to db config} --type=storage --storage-root={state root hex string} --address={contract address hex string}
"`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		validateTrie()
	},
}

func validateTrie() {
	v, err := newValidator()
	if err != nil {
		logWithCommand.Fatal(err)
	}
	switch strings.ToLower(validationType) {
	case "f", "full":
		if stateRootStr == "" {
			logWithCommand.Fatal("must provide a state root for full state validation")
		}
		stateRoot := common.HexToHash(stateRootStr)
		if err = v.ValidateTrie(stateRoot); err != nil {
			logWithCommand.Fatalf("State for root %s is not complete\r\nerr: %v", stateRoot.String(), err)
		}
		logWithCommand.Infof("State for root %s is complete", stateRoot.String())
	case "state":
		if stateRootStr == "" {
			logWithCommand.Fatal("must provide a state root for state trie validation")
		}
		stateRoot := common.HexToHash(stateRootStr)
		if err = v.ValidateStateTrie(stateRoot); err != nil {
			logWithCommand.Fatalf("State trie for root %s is not complete\r\nerr: %v", stateRoot.String(), err)
		}
		logWithCommand.Infof("State trie for root %s is complete", stateRoot.String())
	case "storage":
		if storageRootStr == "" {
			logWithCommand.Fatal("must provide a storage root for storage trie validation")
		}
		if contractAddrStr == "" {
			logWithCommand.Fatal("must provide a contract address for storage trie validation")
		}
		storageRoot := common.HexToHash(storageRootStr)
		addr := common.HexToAddress(contractAddrStr)
		if err = v.ValidateStorageTrie(addr, storageRoot); err != nil {
			logWithCommand.Fatalf("Storage trie for contract %s and root %s not complete\r\nerr: %v", addr.String(), storageRoot.String(), err)
		}
		logWithCommand.Infof("Storage trie for contract %s and root %s is complete", addr.String(), storageRoot.String())
	}
}

func newValidator() (*validator.Validator, error) {
	if ipfsPath == "" {
		db, err := validator.NewDB()
		if err != nil {
			logWithCommand.Fatal(err)
		}
		return validator.NewPGIPFSValidator(db), nil
	}
	bs, err := validator.InitIPFSBlockService(ipfsPath)
	if err != nil {
		return nil, err
	}
	return validator.NewIPFSValidator(bs), nil
}

func init() {
	rootCmd.AddCommand(validateTrieCmd)
	validateTrieCmd.Flags().StringVarP(&stateRootStr, "state-root", "s", "", "Root of the state trie we wish to validate; for full or state validation")
	validateTrieCmd.Flags().StringVarP(&validationType, "type", "t", "full", "Type of validations: full, state, storage")
	validateTrieCmd.Flags().StringVarP(&storageRootStr, "storage-root", "o", "", "Root of the storage trie we wish to validate; for storage validation")
	validateTrieCmd.Flags().StringVarP(&contractAddrStr, "address", "a", "", "Contract address for the storage trie we wish to validate; for storage validation")
	validateTrieCmd.Flags().StringVarP(&ipfsPath, "ipfs-path", "i", "", "Path to IPFS repository")
}
