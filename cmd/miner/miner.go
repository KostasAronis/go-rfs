package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/KostasAronis/go-rfs/blockchain/miner"
	"github.com/KostasAronis/go-rfs/filesystem"
	"github.com/KostasAronis/go-rfs/minerconfig"
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
