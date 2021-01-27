package miner_test

import (
	"log"
	"testing"

	"github.com/KostasAronis/go-rfs/blockchain/miner"
)

func TestMiner(t *testing.T) {
	minerConfig := miner.Config{
		MinerID:             "1",
		PeerMinersAddrs:     []string{"127.0.0.1:9002", "127.0.0.1:9003", "127.0.0.1:9004"},
		IncomingClientsAddr: "127.0.0.1:8001",
		IncomingMinersAddr:  "127.0.0.1:9001",
		OutgoingMinersIP:    "127.0.0.1:10001",
		CommonMinerConfig: miner.CommonMinerConfig{
			GenOpBlockTimeout:      500,
			MinedCoinsPerOpBlock:   3,
			MinedCoinsPerNoOpBlock: 2,
			NumCoinsPerFileCreate:  1,
			PowPerOpBlock:          6,
			PowPerNoOpBlock:        6,
			ConfirmsPerFileCreate:  1,
			ConfirmsPerFileAppend:  2,
			GenesisBlockHash:       "",
		},
	}
	miner := miner.New(&minerConfig)
	log.Println(miner.Coins)
}
