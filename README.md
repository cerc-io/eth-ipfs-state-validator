# eth-ipfs-state-validator

Uses [pg-ipfs-ethdb](https://github.com/vulcanize/pg-ipfs-ethdb) to validate completeness of Ethereum state data on PG-IPFS

## Usage

Run

`./eth-ipfs-state-validator validateTrie --root={state root string} --config={path to .toml config file} `

With `root` as the state root hash we want to validate the corresponding trie for.
The config file holds the parameters for connecting to the IPFS-backing Postgres database.

```toml
[database]
    name     = "vulcanize_public"
    hostname = "localhost"
    user     = "postgres"
    password = ""
    port     = 5432
```