package tcp

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
)

//Client describes a simple tcp client
type Client struct {
	ID          string
	Address     string
	TargetAddr  string
	TargetID    string
	GovecLogger *govec.GoLog
	msgQueue    chan *queuedRequest
}

type queuedRequest struct {
	req      *Msg
	resChan  chan *Msg
	govecTag string
}

//Send Sends tcp message to server
func (c *Client) Send(msg *Msg, govecTag string) *Msg {
	if c.msgQueue == nil {
		c.msgQueue = make(chan *queuedRequest, 20)
		go c.send()
	}
	resChan := make(chan *Msg)
	c.msgQueue <- &queuedRequest{
		req:      msg,
		resChan:  resChan,
		govecTag: govecTag,
	}
	res := <-resChan
	return res
}
func (c *Client) send() {
	for {
		queuedRequest := <-c.msgQueue
		msg := queuedRequest.req
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
		if queuedRequest.govecTag == "" {
			queuedRequest.govecTag = "SendingMessage"
		}
		vectorClockMessage := c.GovecLogger.PrepareSend(queuedRequest.govecTag, msg, govec.GetDefaultLogOptions())
		d := net.Dialer{Timeout: 2 * time.Second}
		conn, err := d.Dial("tcp", c.TargetAddr)
		if err != nil {
			log.Printf("TCP DIAL ERR: %s", err.Error())
			queuedRequest.resChan <- &Msg{
				ClientID: c.ID,
				MSGType:  Error,
				Payload:  err.Error(),
			}
			continue
		}
		//log.Printf("send: %+v\n", vectorClockMessage)
		n, err := conn.Write(vectorClockMessage)
		if err != nil {
			log.Println("TCP WRITE ERR: " + err.Error())
			continue
		}
		resBytes := make([]byte, 4096)
		n, err = conn.Read(resBytes)
		res := Msg{}
		json.Unmarshal(resBytes[0:n], &res)
		//log.Printf("recv: %+v\n", res)
		queuedRequest.resChan <- &res
	}
}
