package tcp

//MSGType enum for labeling operation types
type MSGType int

const (
	//Error some error on the connection
	Error MSGType = 0
	//CreateFile message send by client and peer miners
	CreateFile MSGType = 1
	//AppendRec message send by client and peer miners
	AppendRec MSGType = 2
	//ListFiles message send by client
	ListFiles MSGType = 3
	//TotalRecs message send by client
	TotalRecs MSGType = 4
	//ReadRec message send by client
	ReadRec MSGType = 5
	//Block message send by peer miners
	Block MSGType = 6
	//StoreAndStop stops the server and stores the blockchain to a file
	StoreAndStop = 7
)

func (m MSGType) String() string {
	switch m {
	case Error:
		return "Error"
	case CreateFile:
		return "CreateFile"
	case AppendRec:
		return "AppendRec"
	case ListFiles:
		return "ListFiles"
	case TotalRecs:
		return "TotalRecs"
	case ReadRec:
		return "ReadRec"
	case Block:
		return "Block"
	default:
		return "UnknownMsg"
	}
}
