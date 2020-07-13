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

	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipfs/go-ipfs-ds-help"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
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

// InitIPFSBlockService is used to configure and return a BlockService using an ipfs repo path (e.g. ~/.ipfs)
func InitIPFSBlockService(ipfsPath string) (blockservice.BlockService, error) {
	r, openErr := fsrepo.Open(ipfsPath)
	if openErr != nil {
		return nil, openErr
	}
	ctx := context.Background()
	cfg := &core.BuildCfg{
		Online: false,
		Repo:   r,
	}
	ipfsNode, newNodeErr := core.NewNode(ctx, cfg)
	if newNodeErr != nil {
		return nil, newNodeErr
	}
	return ipfsNode.Blocks, nil
}

// ResetTestDB drops all rows in the test db public.blocks table
func ResetTestDB(db *sqlx.DB) error {
	_, err := db.Exec("DELETE FROM public.blocks")
	return err
}
