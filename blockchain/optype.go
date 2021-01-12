package blockchain

//OpType enum for labeling operation types
type OpType int

const (
	//CreateFile operation on the blockchain
	CreateFile OpType = iota
	//ListFiles operation on the blockchain
	ListFiles
	//TotalRecs operation on the blockchain
	TotalRecs
	//ReadRec operation on the blockchain
	ReadRec
	//AppendRec operation on the blockchain
	AppendRec
)
