package miner

/*
	TODO:
	1) Abstract all types used by Miner into interfaces for easier mocking / testing
	2) Implement ping pong between miners?
*/

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/KostasAronis/go-rfs/blockchain"
	"github.com/KostasAronis/go-rfs/filesystem"
	"github.com/KostasAronis/go-rfs/tcp"
)

const maxNonce uint32 = 4294967295
const goRoutineCount = 5

//Miner describes the main miner entity of the network
type Miner struct {
	Coins             int
	minerConfig       *MinerConfig
	peers             []*tcp.Client
	blockchainServer  *tcp.Server
	clientServer      *tcp.Server
	fs                *filesystem.FileSystem
	blockchain        *blockchain.BlockTree
	pendingOperations []*blockchain.OpRecord
	timer             *time.Timer
}

//New is a makeshift constructor for an initialized (but not started Miner)
func New(minerConfig *MinerConfig) *Miner {
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
		fs: &filesystem.FileSystem{},
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
	m.fs.Init()
	err := m.initBlockchain()
	if err != nil {
		return err
	}
	err = m.startListeningTCP()
	return err
}

func (m *Miner) initBlockchain() error {
	genesisBlock := blockchain.Block{
		PrevHash: "",
		Nonce:    0,
		MinerID:  "0",
		IsOp:     false,
	}
	genesisBlockHash, err := genesisBlock.ComputeHash()
	if err != nil {
		return err
	}
	if m.minerConfig.CommonMinerConfig.GenesisBlockHash != genesisBlockHash {
		return errors.New("incorrect genesis block hash")
	}
	return nil
}

func (m *Miner) startListeningTCP() error {
	blockChainSend, blockChainRecv, clientSend, clientRecv := make(chan *tcp.Msg), make(chan *tcp.Msg), make(chan *tcp.Msg), make(chan *tcp.Msg)
	// blockChainErr, clientErr := make(chan error), make(chan error)
	err := m.blockchainServer.Start(blockChainRecv, blockChainSend)
	if err != nil {
		return err
	}
	log.Printf("miners server listening on: %s\n", m.minerConfig.IncomingMinersAddr)
	err = m.clientServer.Start(clientRecv, clientSend)
	if err != nil {
		return err
	}
	log.Printf("client server listening on: %s\n", m.minerConfig.IncomingClientsAddr)
	for {
		select {
		case msg, ok := <-blockChainRecv:
			if ok {
				log.Printf("blockchain server recv: %s", msg.MSGType)
				res := m.handleBlockchainMsg(msg)
				blockChainSend <- res
			}
		case msg, ok := <-clientRecv:
			if ok {
				res := m.handleClientMsg(msg)
				log.Printf("client server recv: %s", msg.MSGType)
				clientSend <- res
			}
		}
	}
}

func (m *Miner) handleBlockchainMsg(msg *tcp.Msg) *tcp.Msg {
	return nil
}

//TODO: PRETTIFY THIS MESS!!!
func (m *Miner) handleClientMsg(msg *tcp.Msg) *tcp.Msg {
	switch msg.MSGType {
	case tcp.CreateFile:
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
		op := blockchain.OpRecord{
			OpType:   optype,
			MinerID:  m.minerConfig.MinerID,
			Filename: filename.(string),
			Record:   record.(string),
		}
		op.OpType = optype
		m.addOp(&op)
		log.Println(optype)
		return &tcp.Msg{
			MSGType: msg.MSGType,
			Payload: "OpAdded",
		}
	case tcp.ListFiles:
		resPayload := m.fs.ListFiles()
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
		resPayload, err := m.fs.TotalRecords(filename.(string))
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
		resPayload, err := m.fs.ReadRecord(filename.(string), record.(int))
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
			Payload: "!ERROR",
		}
	}
	return nil
}

//TODO: Flood to other peers
func (m *Miner) addOp(op *blockchain.OpRecord) {
	if m.timer == nil {
		m.timer = time.NewTimer(time.Duration(m.minerConfig.CommonMinerConfig.GenOpBlockTimeout) * time.Millisecond)
	}
	m.pendingOperations = append(m.pendingOperations, op)
}

//TODO: handle err, flood finished block to peers
func (m *Miner) mineOpBlock() {
	m.timer = nil
	lastHash, err := m.blockchain.GetLastBlock().ComputeHash()
	if err != nil {
		panic(err)
	}
	block := blockchain.Block{
		MinerID:  m.minerConfig.MinerID,
		PrevHash: lastHash,
		IsOp:     true,
		Ops:      m.pendingOperations,
	}
	m.mine(&block)
	m.blockchain.AppendBlock(&block)
}

//TODO: handle err, flood finished block to peers
func (m *Miner) mineNoopBlock() {
	lastHash, err := m.blockchain.GetLastBlock().ComputeHash()
	if err != nil {
		panic(err)
	}
	block := blockchain.Block{
		MinerID:  m.minerConfig.MinerID,
		PrevHash: lastHash,
		IsOp:     false,
		Ops:      nil,
	}
	m.mine(&block)
	m.blockchain.AppendBlock(&block)
}

//Mine Works on solving a block nonce
func (m *Miner) mine(block *blockchain.Block) {
	var difficulty int
	if block.IsOp {
		difficulty = m.minerConfig.CommonMinerConfig.PowPerOpBlock
	} else {
		difficulty = m.minerConfig.CommonMinerConfig.PowPerNoOpBlock
	}
	nonce := calculateNonceParallel(block, difficulty)
	log.Println(nonce)
	log.Println(block.ComputeHash())
}

func calculateNonceParallel(block *blockchain.Block, difficulty int) uint32 {
	out := make(chan uint32)
	for i := 0; i < goRoutineCount; i++ {
		go tryFindNonce(uint32(i), difficulty, out, *block)
	}
	nonce := <-out
	block.Nonce = nonce
	return nonce
}
func tryFindNonce(goID uint32, difficulty int, out chan uint32, block blockchain.Block) {
	for i := uint32(goID * (maxNonce / goRoutineCount)); i < (goID+1)*(maxNonce/goRoutineCount); i++ {
		block.Nonce = uint32(i)
		valid, err := block.ValidHash(difficulty)
		if err != nil {
			err := fmt.Errorf("Miner: tryFindNonce: error computing hash for block! %s", err.Error())
			log.Println(err)
		}
		if valid {
			out <- block.Nonce
		}
	}
}
