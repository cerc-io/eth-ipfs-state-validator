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
	"github.com/spf13/viper"

	validator "github.com/cerc-io/eth-ipfs-state-validator/v4/pkg"
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
	params := validator.Params{
		Workers:        viper.GetUint("validator.workers"),
		RecoveryFormat: viper.GetString("validator.recoveryFormat"),
	}
	v, err := newValidator(params)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	stateRootStr := viper.GetString("validator.stateRoot")
	storageRootStr := viper.GetString("validator.storageRoot")
	contractAddrStr := viper.GetString("validator.address")
	traversal := strings.ToLower(viper.GetString("validator.type"))
	switch traversal {
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
		if stateRootStr == "" {
			logWithCommand.Fatal("must provide a state root for state trie validation")
		}
		storageRoot := common.HexToHash(storageRootStr)
		addr := common.HexToAddress(contractAddrStr)
		stateRoot := common.HexToHash(stateRootStr)
		if err = v.ValidateStorageTrie(stateRoot, addr, storageRoot); err != nil {
			logWithCommand.Fatalf("Storage trie for contract %s and root %s not complete\r\nerr: %v", addr.String(), storageRoot.String(), err)
		}
		logWithCommand.Infof("Storage trie for contract %s and root %s is complete", addr.String(), storageRoot.String())
	default:
		logWithCommand.Fatalf("Invalid traversal level: '%s'", traversal)
	}

	stats := v.GetCacheStats()
	logWithCommand.Debugf("groupcache stats %+v", stats)
}

func newValidator(params validator.Params) (*validator.Validator, error) {
	ipfsPath := viper.GetString("ipfs.path")
	if ipfsPath == "" {
		db, err := validator.NewDB()
		if err != nil {
			logWithCommand.Fatal(err)
		}
		return validator.NewPGIPFSValidator(db, params), nil
	}
	bs, err := validator.InitIPFSBlockService(ipfsPath)
	if err != nil {
		return nil, err
	}
	return validator.NewIPFSValidator(bs, params), nil
}

func init() {
	rootCmd.AddCommand(validateTrieCmd)

	validateTrieCmd.PersistentFlags().String("state-root", "", "Root of the state trie we wish to validate; for full or state validation")
	validateTrieCmd.PersistentFlags().String("type", "", "Type of validations: full, state, storage")
	validateTrieCmd.PersistentFlags().String("storage-root", "", "Root of the storage trie we wish to validate; for storage validation")
	validateTrieCmd.PersistentFlags().String("address", "", "Contract address for the storage trie we wish to validate; for storage validation")
	validateTrieCmd.PersistentFlags().String("ipfs-path", "", "Path to IPFS repository; if provided operations move through the IPFS repo otherwise Postgres connection params are expected in the provided config")
	validateTrieCmd.PersistentFlags().Int("workers", 4, "number of concurrent workers to use")
	validateTrieCmd.PersistentFlags().String("recovery-format", validator.DefaultRecoveryFormat, "format pattern for recovery files")

	viper.BindPFlag("validator.stateRoot", validateTrieCmd.PersistentFlags().Lookup("state-root"))
	viper.BindPFlag("validator.type", validateTrieCmd.PersistentFlags().Lookup("type"))
	viper.BindPFlag("validator.storageRoot", validateTrieCmd.PersistentFlags().Lookup("storage-root"))
	viper.BindPFlag("validator.address", validateTrieCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("validator.workers", validateTrieCmd.PersistentFlags().Lookup("workers"))
	viper.BindPFlag("validator.recoveryFormat", validateTrieCmd.PersistentFlags().Lookup("recovery-format"))
	viper.BindPFlag("ipfs.path", validateTrieCmd.PersistentFlags().Lookup("ipfs-path"))
}
