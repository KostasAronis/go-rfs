package miner

/*
	TODO:
	1) Abstract all types used by Miner into interfaces for easier mocking / testing
	2) Implement ping pong between miners?
*/

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/KostasAronis/go-rfs/blockchain"
	"github.com/KostasAronis/go-rfs/blockchainfs"
	"github.com/KostasAronis/go-rfs/minerconfig"
	"github.com/KostasAronis/go-rfs/rfslib"
	"github.com/KostasAronis/go-rfs/tcp"
	"github.com/KostasAronis/go-rfs/uuid"
)

const maxNonce uint32 = 4294967295
const goRoutineCount = 5

//Miner describes the main miner entity of the network
type Miner struct {
	bank              map[string]int
	blockchainfs      *blockchainfs.BlockchainFS
	Coins             int
	minerConfig       *minerconfig.Config
	peers             []*tcp.Client
	blockchainServer  *tcp.Server
	clientServer      *tcp.Server
	pendingOperations []*blockchain.OpRecord
	timer             *time.Timer
	OpBlockProduced   chan *blockchain.Block
}

//New is a makeshift constructor for an initialized (but not started Miner)
func New(minerConfig *minerconfig.Config) *Miner {
	m := Miner{
		minerConfig: minerConfig,
		Coins:       0,
		peers:       []*tcp.Client{},
		blockchainServer: &tcp.Server{
			Address: minerConfig.IncomingMinersAddr,
		},
		clientServer: &tcp.Server{
			Address: minerConfig.IncomingClientsAddr,
		},
		blockchainfs: &blockchainfs.BlockchainFS{},
	}
	return &m
}

//Start Starts the miner tcp servers, clients and starts mining for noop blocks
func (m *Miner) Start() error {
	for i := 0; i < len(m.minerConfig.PeerMinersAddrs); i++ {
		c := tcp.Client{
			Address: m.minerConfig.OutgoingMinersIP,
			Target:  m.minerConfig.PeerMinersAddrs[i],
		}
		m.peers = append(m.peers, &c)
	}
	m.blockchainfs = &blockchainfs.BlockchainFS{}
	err := m.blockchainfs.Init(m.minerConfig)
	if err != nil {
		return err
	}
	err = m.startListeningTCP()
	return err
}

func (m *Miner) startListeningTCP() error {
	blockChainConnections, clientConnections := make(chan *tcp.Connection, 10), make(chan *tcp.Connection, 10)
	// blockChainErr, clientErr := make(chan error), make(chan error)
	err := m.blockchainServer.Start(blockChainConnections)
	if err != nil {
		return err
	}
	log.Printf("miners server listening on: %s\n", m.minerConfig.IncomingMinersAddr)
	err = m.clientServer.Start(clientConnections)
	if err != nil {
		return err
	}
	log.Printf("client server listening on: %s\n", m.minerConfig.IncomingClientsAddr)
	for {
		select {
		case conn, ok := <-blockChainConnections:
			if ok {
				go m.handleBlockchainConn(conn)
			}
		case conn, ok := <-clientConnections:
			if ok {
				go m.handleClientConn(conn)
			}
		}
	}
}

func (m *Miner) handleBlockchainConn(conn *tcp.Connection) {
	msg := <-conn.Recv
	log.Printf("blockchain server recv: %s", msg.MSGType)
	conn.Send <- m.handleBlockchainMsg(msg)
}

func (m *Miner) handleClientConn(conn *tcp.Connection) {
	msg := <-conn.Recv
	log.Printf("client server recv: %s", msg.MSGType)
	conn.Send <- m.handleClientMsg(msg)
}

func (m *Miner) handleBlockchainMsg(msg *tcp.Msg) *tcp.Msg {
	switch msg.MSGType {
	//TODO: IMPLEMENT IF NEEDED
	// case tcp.CreateFile:
	// case tcp.AppendRec:
	case tcp.OpBlock:
	case tcp.NoopBlock:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		block := blockchain.Block{}
		blockBytes, ok := payload["Block"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		err := json.Unmarshal(blockBytes.([]byte), &block)
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload: " + err.Error(),
			}
		}
		if block.IsOp {
			if err != nil {
				return &tcp.Msg{
					MSGType: tcp.Error,
					Payload: "op error: " + err.Error(),
				}
			}
		}
		//err = m.appendBlock(&block)
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "block error: " + err.Error(),
			}
		}
		err = m.flood(msg)
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "flood error: " + err.Error(),
			}
		}
	default:
		return &tcp.Msg{
			MSGType: tcp.Error,
			Payload: "NIY",
		}
	}
	return &tcp.Msg{
		MSGType: tcp.Error,
		Payload: "NIY",
	}
}

// func (m *Miner) appendBlock(block *blockchain.Block) error {
// 	err := m.applyBlockOps(block)
// 	err = m.blockchain.AppendBlock(block)
// 	return err
// }

// func (m *Miner) applyBlockOps(block *blockchain.Block) error {
// 	stagingFS := m.fs.Clone()
// 	stagingBank := map[string]int{}
// 	for k, v := range m.bank {
// 		stagingBank[k] = v
// 	}
// 	for _, op := range block.Ops {
// 		if op.OpType == blockchain.CreateFile {
// 			minerCoins, ok := stagingBank[op.MinerID]
// 			if !ok {
// 				return errors.New(op.MinerID + " invalid coin count")
// 			}
// 			if minerCoins-m.minerConfig.CommonMinerConfig.NumCoinsPerFileCreate < 0 {
// 				return errors.New(op.MinerID + " invalid coin count")
// 			}
// 			stagingBank[op.MinerID] = minerCoins - m.minerConfig.CommonMinerConfig.NumCoinsPerFileCreate
// 			_, err := stagingFS.AddFile(op.Filename)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		if op.OpType == blockchain.AppendRec {
// 			_, err := stagingFS.AppendRecord(op.Filename, op.Record)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
func (m *Miner) flood(msg *tcp.Msg) error {
	errors := []string{}
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, client := range m.peers {
		if client.ID != msg.ClientID {
			wg.Add(1)
			go func(c *tcp.Client) {
				res := c.Send(msg)
				if res.MSGType == tcp.Error {
					mutex.Lock()
					defer mutex.Unlock()
					err := res.Payload.(string)
					errors = append(errors, c.Target+": "+err)
					wg.Done()
				}
			}(client)
		}
	}
	wg.Wait()
	return fmt.Errorf(strings.Join(errors, ", "))
}
func (m *Miner) handleClientMsg(msg *tcp.Msg) *tcp.Msg {
	switch msg.MSGType {
	case tcp.CreateFile:
		optype := blockchain.OpType(msg.MSGType)
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		filename, ok := payload["Filename"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		record, ok := payload["Record"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		uuid, err := uuid.New()
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		op := blockchain.OpRecord{
			OpType:   optype,
			MinerID:  m.minerConfig.MinerID,
			Filename: filename.(string),
			Record:   record.(*rfslib.Record),
			UUID:     uuid,
		}
		op.OpType = optype
		blockChan, err := m.blockchainfs.TryStageOp(&op)
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		<-blockChan
		return &tcp.Msg{
			MSGType: msg.MSGType,
			Payload: "OpAdded",
		}

	case tcp.AppendRec:
		optype := blockchain.OpType(msg.MSGType)
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		filename, ok := payload["Filename"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		record, ok := payload["Record"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		uuid, err := uuid.New()
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		op := blockchain.OpRecord{
			OpType:   optype,
			MinerID:  m.minerConfig.MinerID,
			Filename: filename.(string),
			Record:   record.(*rfslib.Record),
			UUID:     uuid,
		}
		op.OpType = optype
		blockChan, err := m.blockchainfs.TryStageOp(&op)
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		<-blockChan
		return &tcp.Msg{
			MSGType: msg.MSGType,
			Payload: "OpAdded",
		}
	case tcp.ListFiles:
		resPayload := m.blockchainfs.FS.ListFiles()
		return &tcp.Msg{
			MSGType: msg.MSGType,
			Payload: resPayload,
		}
	case tcp.TotalRecs:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		filename, ok := payload["Filename"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		resPayload, err := m.blockchainfs.FS.TotalRecords(filename.(string))
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		return &tcp.Msg{
			MSGType: msg.MSGType,
			Payload: resPayload,
		}
	// Read record operation on the rfs, no blocking
	case tcp.ReadRec:
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		filename, ok := payload["Filename"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		record, ok := payload["Record"]
		if !ok {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: "Incorrect payload",
			}
		}
		resPayload, err := m.blockchainfs.FS.ReadRecord(filename.(string), record.(int))
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		return &tcp.Msg{
			MSGType: msg.MSGType,
			Payload: resPayload,
		}
	case tcp.Error:
	default:
		return &tcp.Msg{
			MSGType: tcp.Error,
			Payload: "Incorrect MSGType",
		}
	}
	return &tcp.Msg{
		MSGType: tcp.Error,
		Payload: "Incorrect MSGType",
	}
}
