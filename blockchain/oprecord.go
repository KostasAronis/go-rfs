package blockchain

//OpRecord describes an operation on the rfs stored on the blockchain
type OpRecord struct {
	OpType  OpType
	MinerID string
}
