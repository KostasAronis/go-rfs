package blockchain

import (
	"time"

	"github.com/KostasAronis/go-rfs/rfslib"
)

//OpRecord describes an operation on the rfs stored on the blockchain
type OpRecord struct {
	MinerID   string
	Timestamp time.Time
	UUID      string
	OpType    OpType
	Filename  string
	Record    *rfslib.Record
}
