package tcp

//MSGType enum for labeling operation types
type MSGType int

const (
	//Read operation on the blockchain
	Read MSGType = iota
	//Write operation on the blockchain
	Write
)

type TCPMsg struct {
	MSGType MSGType
	Payload interface{}
}
