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
	"context"

	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql/postgres"
	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/spf13/viper"
)

// Env variables
const (
	DATABASE_NAME     = "DATABASE_NAME"
	DATABASE_HOSTNAME = "DATABASE_HOSTNAME"
	DATABASE_PORT     = "DATABASE_PORT"
	DATABASE_USER     = "DATABASE_USER"
	DATABASE_PASSWORD = "DATABASE_PASSWORD"
)

// NewDB returns a new sqlx.DB from config/cli/env variables
func NewDB(ctx context.Context) (sql.Database, error) {
	var c postgres.Config
	Init(&c)
	info := node.Info{
		GenesisBlock: "GenesisBlock",
		NetworkID:    "1",
		ChainID:      1,
		ID:           "1",
		ClientName:   "geth",
	}

	driver, err := postgres.NewPGXDriver(ctx, c,info)
	if err != nil {
		return nil, err
	}

	return postgres.NewPostgresDB(driver), nil
}

func Init(c *postgres.Config) {
	viper.BindEnv("database.name", DATABASE_NAME)
	viper.BindEnv("database.hostname", DATABASE_HOSTNAME)
	viper.BindEnv("database.port", DATABASE_PORT)
	viper.BindEnv("database.user", DATABASE_USER)
	viper.BindEnv("database.password", DATABASE_PASSWORD)

	c.DatabaseName = viper.GetString("database.name")
	c.Hostname = viper.GetString("database.hostname")
	c.Port = viper.GetInt("database.port")
	c.Username = viper.GetString("database.user")
	c.Password = viper.GetString("database.password")
}
