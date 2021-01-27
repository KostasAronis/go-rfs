package tcp

type Msg struct {
	ClientID string
	MSGType  MSGType
	Payload  interface{}
}
