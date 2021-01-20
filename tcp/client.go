package tcp

import (
	"encoding/json"
	"log"
	"net"
)

//Client describes a simple tcp client
type Client struct {
	Address string
	Target  string
}

//Send Sends tcp message to server
func (c *Client) Send(msg *Msg) *Msg {
	conn, err := net.Dial("tcp", c.Target)
	if err != nil {
		log.Println(err)
	}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
	}
	log.Printf("send: %+v\n", msg)
	n, err := conn.Write(b)
	if err != nil {
		log.Println(err)
	}
	resBytes := make([]byte, 4096)
	n, err = conn.Read(resBytes)
	res := Msg{}
	json.Unmarshal(resBytes[0:n], &res)
	log.Printf("recv: %+v\n", res)
	return &res
}
