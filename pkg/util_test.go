package validator_test

import (
	"github.com/ipfs/go-cid"
	"github.com/jmoiron/sqlx"
)

// PublishRaw derives a cid from raw bytes and provided codec and multihash type, and writes it to the db tx
func PublishRaw(tx *sqlx.Tx, codec, mh uint64, raw []byte, blockNumber uint64) error {
	c, err := RawdataToCid(codec, raw, mh)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`INSERT INTO ipld.blocks (key, data, block_number) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		c.String(), raw, blockNumber)
	return err
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

// ResetTestDB truncates all used tables from the test DB
func ResetTestDB(db *sqlx.DB) error {
	_, err := db.Exec("TRUNCATE ipld.blocks")
	return err
}
