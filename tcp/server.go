package tcp

import (
	"encoding/json"
	"log"
	"net"
)

//Server describes a server listening for tcp messages from tcp clients
type Server struct {
	Address string
}

type Connection struct {
	Recv chan *Msg
	Send chan *Msg
}

//Start Starts the server and listens for tcp messages
func (s *Server) Start(connectionsChannel chan *Connection) error {
	go func() {
		l, err := net.Listen("tcp4", s.Address)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		defer l.Close()
		for {
			c, err := l.Accept()
			if err != nil {
				log.Println(err)
				panic(err)
			}
			cChan := Connection{
				Recv: make(chan *Msg),
				Send: make(chan *Msg),
			}
			go s.waitForResponse(c.(*net.TCPConn), &cChan)
			connectionsChannel <- &cChan
		}
	}()
	return nil
}
func (s *Server) waitForResponse(c *net.TCPConn, conn *Connection) {
	data := make([]byte, 2048)
	n, err := c.Read(data)
	if err != nil {
		log.Println(err)
	}
	msg := Msg{}
	err = json.Unmarshal(data[0:n], &msg)
	conn.Recv <- &msg
	response := <-conn.Send
	resBytes, err := json.Marshal(response)
	c.Write(resBytes)
}
