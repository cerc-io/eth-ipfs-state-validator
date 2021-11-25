module github.com/vulcanize/eth-ipfs-state-validator

go 1.13

require (
	github.com/ethereum/go-ethereum v1.10.11
	github.com/ipfs/go-blockservice v0.1.7
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-filestore v1.0.0 //indirect
	github.com/ipfs/go-ipfs v0.10.0
	github.com/ipfs/go-ipfs-blockstore v1.0.1
	github.com/ipfs/go-ipfs-ds-help v1.0.0
	github.com/lib/pq v1.10.2
	github.com/mailgun/groupcache/v2 v2.2.1
	github.com/multiformats/go-multihash v0.0.15
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/vulcanize/ipfs-ethdb v0.0.5
)

replace (
	github.com/ethereum/go-ethereum v1.10.11 => github.com/Vulcanize/go-ethereum v0.0.0-20211125055606-cd7c58e7f9a2
	github.com/vulcanize/ipfs-ethdb v0.0.5 => github.com/Vulcanize/ipfs-ethdb v0.0.0-20211125060829-0aa16344859a
)
