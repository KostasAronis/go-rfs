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

//Start Starts the server and listens for tcp messages
func (s *Server) Start(recv chan *Msg, send chan *Msg) error {
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
			data := make([]byte, 2048)
			n, err := c.Read(data)
			if err != nil {
				log.Println(err)
			}
			msg := Msg{}
			err = json.Unmarshal(data[0:n], &msg)
			recv <- &msg
			response := <-send
			resBytes, err := json.Marshal(response)
			c.Write(resBytes)
		}
	}()
	return nil
}
