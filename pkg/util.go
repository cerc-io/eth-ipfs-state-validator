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
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/repo/fsrepo"
)

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
