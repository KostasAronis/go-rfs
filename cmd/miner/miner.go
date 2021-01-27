package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/KostasAronis/go-rfs/blockchain"
	"github.com/KostasAronis/go-rfs/blockchain/miner"
	"github.com/KostasAronis/go-rfs/filesystem"
	"github.com/KostasAronis/go-rfs/minerconfig"
	"github.com/KostasAronis/go-rfs/tcp"
)

var fs filesystem.FileSystem

func init() {
}
func main() {
	configFile, _ := ioutil.ReadFile("config.json")
	config := minerconfig.Config{}
	err := json.Unmarshal([]byte(configFile), &config)
	if err != nil {
		panic(err)
	}
	log.SetPrefix("Miner " + config.MinerID + ": ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmsgprefix | log.Lshortfile)
	m := miner.New(&config)
	err = m.Start()
	if err != nil {
		panic(err)
	}
}

func handleClientMsg(msg *tcp.Msg) *tcp.Msg {
	if msg.MSGType == tcp.MSGType(blockchain.AppendRec) || msg.MSGType == tcp.MSGType(blockchain.CreateFile) {
		optype := blockchain.OpType(msg.MSGType)
		log.Println(optype)
	}
	log.Println("not an op")
	return nil
}
