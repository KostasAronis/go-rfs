package blockchain

import "github.com/KostasAronis/go-rfs/rfslib"

//OpRecord describes an operation on the rfs stored on the blockchain
type OpRecord struct {
	UUID     string
	OpType   OpType
	MinerID  string
	Filename string
	Record   *rfslib.Record
}
