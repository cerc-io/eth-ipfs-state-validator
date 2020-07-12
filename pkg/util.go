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
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipfs/go-ipfs-ds-help"
	"github.com/jmoiron/sqlx"
)

// PublishRaw derives a cid from raw bytes and provided codec and multihash type, and writes it to the db tx
func PublishRaw(tx *sqlx.Tx, codec, mh uint64, raw []byte) (string, error) {
	c, err := RawdataToCid(codec, raw, mh)
	if err != nil {
		return "", err
	}
	dbKey := dshelp.MultihashToDsKey(c.Hash())
	prefixedKey := blockstore.BlockPrefix.String() + dbKey.String()
	_, err = tx.Exec(`INSERT INTO public.blocks (key, data) VALUES ($1, $2) ON CONFLICT (key) DO NOTHING`, prefixedKey, raw)
	return c.String(), err
}

// RawdataToCid takes the desired codec, multihash type, and a slice of bytes
// and returns the proper cid of the object.
func RawdataToCid(codec uint64, rawdata []byte, multiHash uint64) (cid.Cid, error) {
	c, err := cid.Prefix{
		Codec:    codec,
		Version:  1,
		MhType:   multiHash,
		MhLength: -1,
	}.Sum(rawdata)
	if err != nil {
		return cid.Cid{}, err
	}
	return c, nil
}
