package blockchain

import "github.com/KostasAronis/go-rfs/tcp"

//OpType enum for labeling operation types
type OpType int

const (
	//CreateFile operation on the blockchain
	CreateFile OpType = OpType(tcp.CreateFile)
	//AppendRec operation on the blockchain
	AppendRec OpType = OpType(tcp.AppendRec)
)

func (t OpType) String() string {
	switch t {
	case CreateFile:
		return "CreateFile"
	case AppendRec:
		return "AppendRec"
	default:
		return "UnknownMsg"
	}
}
