package blockchain

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
)

/*
	TODO: Convert Ops to a more generic IContent interface for things other than fs operations to be stored in the blockchain
*/

//Block contains multiple operations on the rfs
type Block struct {
	/*
		PrevHash the hash of the previous block in the chain.
		Must be MD5 hash and contain {config.Difficulty} number of zeroes at the end of the hex representation.
	*/
	PrevHash string
	//Hash stores the md5 hash of the current block
	hash string
	// MinerID the id of the miner that computed this block
	MinerID string
	Nonce   uint32
	//IsOp identifies between Operation and NoOperation blocks
	IsOp bool
	//Ops An ordered set of operation records
	Ops []*OpRecord
}

func (b *Block) GetComputedHash() (string, error) {
	return b.hash, nil
}
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

func (b *Block) ValidHash(difficulty int) (bool, error) {
	hash, err := b.ComputeHash()
	if err != nil {
		return false, err
	}
	return validHash(hash, difficulty), nil
}

func validHash(hash string, difficulty int) bool {
	lastNDigits := hash[len(hash)-int(difficulty):]
	nZeros := strings.Repeat("0", int(difficulty))
	if lastNDigits == nZeros {
		return true
	}
	return false
}
