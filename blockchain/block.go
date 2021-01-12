package blockchain

//Block contains multiple operations on the rfs
type Block struct {
	/*
		PrevHash the hash of the previous block in the chain.
		Must be MD5 hash and contain {config.Difficulty} number of zeroes at the end of the hex representation.
	*/
	PrevHash string
	// MinerID the id of the miner that computed this block
	MinerID string
	Nonce   int32
	//IsOp identifies between Operation and NoOperation blocks
	IsOp bool
	//Ops An ordered set of operation records
	Ops  []OpRecord
	Hash string
}

var GenesisBlock Block = Block{
	PrevHash: "",
	MinerID:  "",
	IsOp:     false,
	Ops:      []OpRecord{},
	Hash:     "",
	Nonce:    0,
}
