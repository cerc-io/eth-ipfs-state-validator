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
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //postgres driver

	"github.com/vulcanize/ipfs-blockchain-watcher/pkg/config"
)

// NewDB returns a new sqlx.DB from env variables
func NewDB() (*sqlx.DB, error) {
	c := config.Database{}
	c.Init()
	connectStr := config.DbConnectionString(c)
	return sqlx.Connect("postgres", connectStr)
}
