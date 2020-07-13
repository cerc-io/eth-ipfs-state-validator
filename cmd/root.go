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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	subCommand      string
	logWithCommand  logrus.Entry
	stateRootStr    string
	storageRootStr  string
	validationType  string
	contractAddrStr string
	cfgFile         string
	ipfsPath        string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "eth-ipfs-state-validator",
	PersistentPreRun: initFuncs,
}

// Execute begins execution of the command
func Execute() {
	logrus.Info("----- Starting eth-ipfs-state-validator -----")
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func initFuncs(cmd *cobra.Command, args []string) {
	logfile := viper.GetString("logfile")
	if logfile != "" {
		file, err := os.OpenFile(logfile,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logrus.Infof("Directing output to %s", logfile)
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
			logrus.Info("Failed to log to file, using default stdout")
		}
	} else {
		logrus.SetOutput(os.Stdout)
	}
	if err := logLevel(); err != nil {
		logrus.Fatal("Could not set log level: ", err)
	}
}

func logLevel() error {
	lvl, err := logrus.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	if lvl > logrus.InfoLevel {
		logrus.SetReportCaller(true)
	}
	logrus.Info("Log level set to ", lvl.String())
	return nil
}

func init() {
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file location")
	rootCmd.PersistentFlags().String("logfile", "", "file path for logging")
	rootCmd.PersistentFlags().String("database-name", "vulcanize_public", "database name")
	rootCmd.PersistentFlags().Int("database-port", 5432, "database port")
	rootCmd.PersistentFlags().String("database-hostname", "localhost", "database hostname")
	rootCmd.PersistentFlags().String("database-user", "", "database user")
	rootCmd.PersistentFlags().String("database-password", "", "database password")
	rootCmd.PersistentFlags().String("log-level", logrus.InfoLevel.String(), "Log level (trace, debug, info, warn, error, fatal, panic")

	viper.BindPFlag("logfile", rootCmd.PersistentFlags().Lookup("logfile"))
	viper.BindPFlag("database.name", rootCmd.PersistentFlags().Lookup("database-name"))
	viper.BindPFlag("database.port", rootCmd.PersistentFlags().Lookup("database-port"))
	viper.BindPFlag("database.hostname", rootCmd.PersistentFlags().Lookup("database-hostname"))
	viper.BindPFlag("database.user", rootCmd.PersistentFlags().Lookup("database-user"))
	viper.BindPFlag("database.password", rootCmd.PersistentFlags().Lookup("database-password"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
}
