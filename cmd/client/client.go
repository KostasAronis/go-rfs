package main

import (
	"log"

	"github.com/KostasAronis/go-rfs/tcp"
)

// TODO: flesh out, currently just for testing
func main() {
	// cc := make(chan int)
	// ccc := make(chan int)
	// go write(cc)
	// for {
	// 	select {
	// 	case i := <-cc:
	// 		log.Println(i)
	// 	case i := <-ccc:
	// 		log.Println(i)
	// 	default:
	// 		log.Println("nthn yet")
	// 	}
	// }

	c := tcp.Client{
		Address: "who cares",
		Target:  ":8001",
	}
	res := c.Send(&tcp.Msg{
		MSGType: tcp.AppendRec,
		Payload: map[string]interface{}{
			"Filename": "test.txt",
			"Record":   "Lorem Ipsum",
		},
	})
	log.Println(res)
}
func write(c chan int) {
	log.Println("writting")
	c <- 5
}
