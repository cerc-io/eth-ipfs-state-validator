# eth-ipfs-state-validator

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/eth-ipfs-state-validator)](https://goreportcard.com/report/github.com/vulcanize/eth-ipfs-state-validator)

> Uses [ipfs-ethdb](https://github.com/vulcanize/ipfs-ethdb/tree/master/postgres) to validate completeness of IPFS Ethereum state data

## Background

State data on Ethereum takes the form of [Modified Merkle Patricia Tries](https://eth.wiki/en/fundamentals/patricia-tree).
On disk each unique node of a trie is stored as a key-value pair between the Keccak256 hash of the RLP-encoded node and the RLP-encoded node.
To prove the existence of a specific node in an MMPT with a known root hash, one provides a list of all of the nodes along the path descending
from the root node to the node in question. To validate the completeness of a state database- to confirm every node for a state and/or storage trie(s) is present
in a database- requires traversing the entire trie (or linked set of tries) and confirming the presence of every node in the database.


## Usage


`full` validates completeness of the entire state corresponding to a provided state root, including both state and storage tries

`./eth-ipfs-state-validator validateTrie --ipfs-path={path to ipfs repo} --type=full --state-root={state root hex string}`


`state` validates completeness of the state trie corresponding to a provided state root, excluding the storage tries

`./eth-ipfs-state-validator validateTrie --ipfs-path={path to ipfs repo} --type=state --state-root={state root hex string}`


`storage` validates completeness of only the storage trie corresponding to a provided storage root and contract address

`./eth-ipfs-state-validator validateTrie --ipfs-path={path to ipfs repo} --type=storage --storage-root={state root hex string} --address={contract address hex string}`


If an IPFS path is provided with the `--ipfs-path` flag, the validator operates through an IPFS block-service and expects a configured IPFS repository at
the provided path. In this case, the validator will vie for contention on the lockfile located at the ipfs path.

Alternatively, if no IPFS path is provided, the `--config` flag can be used to provide a path to a .toml config file with
Postgres database connection parameters. In this case, the validator interfaces directly with the Postgres database and the
database is assumed to be [IPFS-backing](https://github.com/ipfs/go-ds-sql).

Postgres DB config: 

```toml
[database]
    name     = "vulcanize_public"
    hostname = "localhost"
    user     = "postgres"
    password = ""
    port     = 5432
```

## Maintainers
@vulcanize
@AFDudley
@i-norden

## Contributing
Contributions are welcome!

VulcanizeDB follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/1/4/code-of-conduct).

## License
[AGPL-3.0](LICENSE) © Vulcanize Inc