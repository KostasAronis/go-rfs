package tcp

import (
	"encoding/json"
	"fmt"
	"net"
)

type Server struct {
	Address string
}

func (s Server) Start() {
	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		data := []byte{}
		n, err := c.Read(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := TCPMsg{}
		err = json.Unmarshal(data[0:n], &msg)
		response := getResponse(msg)
		resBytes, err := json.Marshal(response)
		c.Write(resBytes)
	}
}

func getResponse(msg TCPMsg) TCPMsg {
	res := TCPMsg{}
	switch res.MSGType {
	case Read:
		return res

	case Write:
		return res

	default:
		return res
	}
}
