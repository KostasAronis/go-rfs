package main

import (
	"github.com/KostasAronis/go-rfs/tcp"
)

// TODO: flesh out, currently just for testing
func main() {
	c := tcp.Client{
		Address: "who cares",
		Target:  ":8001",
	}
	c.Send(&tcp.Msg{
		MSGType: tcp.AppendRec,
		Payload: map[string]interface{}{
			"Filename": "test.txt",
			"Record":   "Lorem Ipsum",
		},
	})
}
