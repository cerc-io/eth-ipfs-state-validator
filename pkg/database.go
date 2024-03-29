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
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
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

type Config struct {
	Hostname string
	Name     string
	User     string
	Password string
	Port     int
}

// NewDB returns a new sqlx.DB from config/cli/env variables
func NewDB() (*sqlx.DB, error) {
	c := Config{}
	LoadViper(&c)
	return sqlx.Connect("postgres", c.ConnString())
}

func (c *Config) ConnString() string {
	if len(c.User) > 0 && len(c.Password) > 0 {
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			c.User, c.Password, c.Hostname, c.Port, c.Name)
	}
	if len(c.User) > 0 && len(c.Password) == 0 {
		return fmt.Sprintf("postgresql://%s@%s:%d/%s?sslmode=disable",
			c.User, c.Hostname, c.Port, c.Name)
	}
	return fmt.Sprintf("postgresql://%s:%d/%s?sslmode=disable", c.Hostname, c.Port, c.Name)
}

func LoadEnv(c *Config) error {
	if val := os.Getenv(DATABASE_NAME); val != "" {
		c.Name = val
	}
	if val := os.Getenv(DATABASE_HOSTNAME); val != "" {
		c.Hostname = val
	}
	if val := os.Getenv(DATABASE_PORT); val != "" {
		port, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		c.Port = port
	}
	if val := os.Getenv(DATABASE_USER); val != "" {
		c.User = val
	}
	if val := os.Getenv(DATABASE_PASSWORD); val != "" {
		c.Password = val
	}
	return nil
}

func LoadViper(c *Config) {
	viper.BindEnv("database.name", DATABASE_NAME)
	viper.BindEnv("database.hostname", DATABASE_HOSTNAME)
	viper.BindEnv("database.port", DATABASE_PORT)
	viper.BindEnv("database.user", DATABASE_USER)
	viper.BindEnv("database.password", DATABASE_PASSWORD)

	c.Name = viper.GetString("database.name")
	c.Hostname = viper.GetString("database.hostname")
	c.Port = viper.GetInt("database.port")
	c.User = viper.GetString("database.user")
	c.Password = viper.GetString("database.password")
}
