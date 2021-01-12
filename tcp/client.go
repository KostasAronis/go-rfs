package tcp

import (
	"encoding/json"
	"fmt"
	"net"
)

type Client struct {
	Address string
	Target  string
}

func (c Client) Send(msg TCPMsg) TCPMsg {
	conn, err := net.Dial("tcp", c.Target)
	if err != nil {
		fmt.Println(err)
	}
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	conn.Write(b)
	resBytes := []byte{}
	n, err := conn.Read(resBytes)
	res := TCPMsg{}
	json.Unmarshal(resBytes[0:n], &res)
	return res
}
