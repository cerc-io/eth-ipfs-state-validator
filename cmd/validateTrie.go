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
	"fmt"

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
	Long: `This command is used to validate the completeness of the state trie corresponding to a specific state root`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		validateTrie()
	},
}

func validateTrie() {
	db, err := validator.NewDB()
	if err != nil {
		logWithCommand.Fatal(err)
	}
	v := validator.NewValidator(db)
	rootHash := common.HexToHash(rootStr)
	if _, err = v.ValidateTrie(rootHash); err != nil {
		fmt.Printf("State trie is not complete\r\nerr: %v", err)
		logWithCommand.Fatal(err)
	}
	fmt.Printf("State trie for root %s is complete", rootStr)
}

func init() {
	rootCmd.AddCommand(validateTrieCmd)
	validateTrieCmd.Flags().StringVarP(&rootStr, "root", "r", "", "Root of the state trie we wish to validate")
}
