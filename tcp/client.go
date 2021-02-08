package tcp

import (
	"encoding/json"
	"log"
	"net"

	"github.com/DistributedClocks/GoVector/govec"
)

//Client describes a simple tcp client
type Client struct {
	ID          string
	Address     string
	TargetAddr  string
	TargetID    string
	GovecLogger *govec.GoLog
}

//Send Sends tcp message to server
func (c *Client) Send(msg *Msg) *Msg {
	log.Printf("sending tcp to %s", c.TargetAddr)
	if msg.ClientID == "" {
		msg.ClientID = c.ID
	}
	if c.GovecLogger == nil {
		goVecConfig := govec.GetDefaultConfig()
		goVecConfig.UseTimestamps = true
		goVecConfig.AppendLog = true
		c.GovecLogger = govec.InitGoVector(c.ID, c.ID+"GoVector.log", goVecConfig)
	}
	vectorClockMessage := c.GovecLogger.PrepareSend("SendingMessage", msg, govec.GetDefaultLogOptions())
	conn, err := net.Dial("tcp", c.TargetAddr)
	if err != nil {
		log.Println(err)
		return &Msg{
			ClientID: c.ID,
			MSGType:  Error,
			Payload:  err.Error(),
		}
	}
	//log.Printf("send: %+v\n", vectorClockMessage)
	n, err := conn.Write(vectorClockMessage)
	if err != nil {
		log.Println(err)
	}
	resBytes := make([]byte, 4096)
	n, err = conn.Read(resBytes)
	res := Msg{}
	json.Unmarshal(resBytes[0:n], &res)
	//log.Printf("recv: %+v\n", res)
	return &res
}
