package blockchain

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
)

/*
	TODO:
	1) Convert Ops to a more generic IContent interface for things other than fs operations to be stored in the blockchain
	2) Add Timestamp
*/

//Block contains multiple operations on the rfs
type Block struct {
	/*
		PrevHash the hash of the previous block in the chain.
		Must be MD5 hash and contain {config.Difficulty} number of zeroes at the end of the hex representation.
	*/
	PrevHash string
	// MinerID the id of the miner that computed this block
	MinerID string
	Nonce   uint32
	//IsOp identifies between Operation and NoOperation blocks
	IsOp bool
	//Ops An ordered set of operation records
	Ops []*OpRecord
	//Confirmations A counter of network confirmations for this block
	Confirmations int
}

//ComputeHash computes the current hash
func (b *Block) ComputeHash() (string, error) {
	// buf := &bytes.Buffer{}
	// err := binary.Write(buf, binary.LittleEndian, b)
	bytes, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	h := md5.New()
	_, err = h.Write(bytes)
	if err != nil {
		return "", err
	}
	str := hex.EncodeToString(h.Sum(nil))
	return str, nil
}

//IsValid computes the current hash, checks it against stored hash and also checks the POW of the block
func (b *Block) IsValid(difficulty int) (bool, error) {
	hash, err := b.ComputeHash()
	if err != nil {
		return false, err
	}
	return validPOW(hash, difficulty), nil
}

//HasValidNonce computes the current hash and checks the pow of the block
func (b *Block) HasValidNonce(difficulty int) (bool, error) {
	hash, err := b.ComputeHash()
	if err != nil {
		return false, err
	}
	return validPOW(hash, difficulty), nil
}

//ValidPOW checks just the last {difficulty} characters of the hash
func validPOW(hash string, difficulty int) bool {
	lastNDigits := hash[len(hash)-int(difficulty):]
	nZeros := strings.Repeat("0", int(difficulty))
	if lastNDigits == nZeros {
		return true
	}
	return false
}
