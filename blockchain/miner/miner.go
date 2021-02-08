package miner

/*
	TODO:
	1) Abstract all types used by Miner into interfaces for easier mocking / testing
	2) Implement ping pong between miners?
*/

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
	"github.com/KostasAronis/go-rfs/blockchain"
	"github.com/KostasAronis/go-rfs/blockchainfs"
	"github.com/KostasAronis/go-rfs/minerconfig"
	"github.com/KostasAronis/go-rfs/rfslib"
	"github.com/KostasAronis/go-rfs/serialization"
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
	blockToFlood      chan *blockchain.Block
	opToFlood         chan *blockchain.OpRecord
	exitError         chan error
}

//New is a makeshift constructor for an initialized (but not started Miner)
func New(minerConfig *minerconfig.Config) *Miner {
	govecConfig := govec.GetDefaultConfig()
	govecConfig.UseTimestamps = true
	govecConfig.AppendLog = true
	govecLogger := govec.InitGoVector(minerConfig.MinerID, minerConfig.MinerID+"GoVector.log", govecConfig)
	blockToFlood := make(chan *blockchain.Block, 10)
	m := Miner{
		minerConfig: minerConfig,
		Coins:       0,
		peers:       []*tcp.Client{},
		blockchainServer: &tcp.Server{
			ID:          minerConfig.MinerID,
			Address:     minerConfig.IncomingMinersAddr,
			GovecLogger: govecLogger,
		},
		clientServer: &tcp.Server{
			ID:          minerConfig.MinerID,
			Address:     minerConfig.IncomingClientsAddr,
			GovecLogger: govecLogger,
		},
		blockchainfs: &blockchainfs.BlockchainFS{
			BlockToFlood: blockToFlood,
			GovecLogger:  govecLogger,
		},
		blockToFlood: blockToFlood,
		opToFlood:    make(chan *blockchain.OpRecord),
	}
	for i := 0; i < len(m.minerConfig.PeerMinersAddrs); i++ {
		c := tcp.Client{
			ID:          m.minerConfig.MinerID,
			Address:     m.minerConfig.OutgoingMinersIP,
			TargetID:    string(m.minerConfig.PeerMinersAddrs[i][len(m.minerConfig.PeerMinersAddrs[i])-1]),
			TargetAddr:  m.minerConfig.PeerMinersAddrs[i],
			GovecLogger: govecLogger,
		}
		m.peers = append(m.peers, &c)
	}
	return &m
}

//Start Starts the miner tcp servers, clients and starts mining for noop blocks
func (m *Miner) Start() error {
	m.exitError = make(chan error)
	go m.floodToPeers()
	err := m.blockchainfs.Init(m.minerConfig)
	if err != nil {
		return err
	}
	err = m.startListeningTCP()
	if err != nil {
		return err
	}
	err = <-m.exitError
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
	go func() {
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
	}()
	return nil
}

func (m *Miner) handleBlockchainConn(conn *tcp.Connection) {
	msg := <-conn.Recv
	log.Printf("blockchain server recv: %s", msg.MSGType)
	conn.Send <- m.handleBlockchainMsg(msg)
}

func (m *Miner) handleBlockchainMsg(msg *tcp.Msg) *tcp.Msg {
	log.Println("got Blockchain msg")
	switch msg.MSGType {
	case tcp.Block:
		blockBytes, ok := msg.Payload.([]byte)
		if !ok {
			return incorrectPayload()
		}
		block, err := serialization.DecodeToBlock(blockBytes)
		if err != nil {
			return errorPayload(err)
		}
		h, err := block.ComputeHash()
		if err != nil {
			return errorPayload(err)
		}
		if m.blockchainfs.BlockExists(h) {
			return msg
		}
		err = m.blockchainfs.AddExternalBlock(block)
		if err != nil {
			return errorPayload(err)
		}
		return msg
	case tcp.CreateFile, tcp.AppendRec:
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

func (m *Miner) handleClientConn(conn *tcp.Connection) {
	msg := <-conn.Recv
	log.Printf("client server recv: %s", msg.MSGType)
	conn.Send <- m.handleClientMsg(msg)
}
func incorrectPayload() *tcp.Msg {
	return &tcp.Msg{
		MSGType: tcp.Error,
		Payload: "Incorrect payload",
	}
}
func errorPayload(err error) *tcp.Msg {
	return &tcp.Msg{
		MSGType: tcp.Error,
		Payload: err.Error(),
	}
}
func (m *Miner) handleClientMsg(msg *tcp.Msg) *tcp.Msg {
	switch msg.MSGType {
	case tcp.CreateFile, tcp.AppendRec:
		optype := blockchain.OpType(msg.MSGType)
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return incorrectPayload()
		}
		filename, ok := payload["Filename"]
		if !ok {
			return incorrectPayload()
		}
		record, ok := payload["Record"]
		if !ok {
			return incorrectPayload()
		}
		uuid, err := uuid.New()
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		r := rfslib.Record{}
		r.FromFloatArrayInterface(record)
		op := blockchain.OpRecord{
			OpType:    optype,
			MinerID:   m.minerConfig.MinerID,
			Filename:  filename.(string),
			Record:    &r,
			UUID:      uuid,
			Timestamp: time.Now().UTC(),
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

	//DEBUG MSG ONLY! CLIENT SHOULDNT CONTROL MINER!
	case tcp.StoreAndStop:
		timestamp := time.Now().Format(time.RFC3339)
		baseFilename := m.clientServer.ID + "_" + timestamp + ".dat"
		filename := getUnusedFilenameToStore(baseFilename)
		err := m.blockchainfs.Store(filename)
		if err != nil {
			return &tcp.Msg{
				MSGType: tcp.Error,
				Payload: err.Error(),
			}
		}
		go func() {
			time.After(500 * time.Millisecond)
			m.exitError <- nil
		}()
		return &tcp.Msg{
			MSGType: tcp.StoreAndStop,
			Payload: map[string]interface{}{
				"Filename": filename,
			},
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
			return incorrectPayload()
		}
		filename, ok := payload["Filename"]
		if !ok {
			return incorrectPayload()
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
			return incorrectPayload()
		}
		filename, ok := payload["Filename"]
		if !ok {
			return incorrectPayload()
		}
		record, ok := payload["Record"]
		if !ok {
			return incorrectPayload()
		}
		indInterfaces, ok := record.([]interface{})
		if !ok {
			return incorrectPayload()
		}
		indexes := []int8{}
		for _, v := range indInterfaces {
			indexes = append(indexes, v.(int8))
		}
		resPayload := []*rfslib.Record{}
		for index := range indexes {
			record, err := m.blockchainfs.FS.ReadRecord(filename.(string), index)
			if err != nil {
				return &tcp.Msg{
					MSGType: tcp.Error,
					Payload: err.Error(),
				}
			}
			resPayload = append(resPayload, record)
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

func (m *Miner) floodToPeers() {
	log.Println("listening for flood msg")
	for {
		log.Println("listening for flood msg_START")
		select {
		case block := <-m.blockToFlood:
			log.Println("got block to flood")
			m.floodBlock(block)
			log.Println("flooded block")
		case op := <-m.opToFlood:
			m.floodOp(op)
		}
		log.Println("listening for flood msg_END")
	}
}

func (m *Miner) floodBlock(block *blockchain.Block) {
	log.Println("flooding block to peers")
	clientsToSend := []*tcp.Client{}
	for _, c := range m.peers {
		if c.TargetID != block.MinerID {
			clientsToSend = append(clientsToSend, c)
		}
	}
	blockBytes, err := serialization.EncodeToBytes(block)
	if err != nil {
		log.Printf("error in encoding: %s", err.Error())
		return
	}
	msg := tcp.Msg{
		ClientID: m.minerConfig.MinerID,
		MSGType:  tcp.Block,
		Payload:  blockBytes,
	}
	err = m.flood(&msg, clientsToSend)
	if err != nil {
		log.Printf("error in flooding: %s", err.Error())
	}
}
func (m *Miner) floodOp(op *blockchain.OpRecord) {
	log.Println("flooding block to peers")
	clientsToSend := []*tcp.Client{}
	for _, c := range m.peers {
		if c.TargetID != op.MinerID {
			clientsToSend = append(clientsToSend, c)
		}
	}
	msg := tcp.Msg{
		ClientID: m.minerConfig.MinerID,
		MSGType:  tcp.MSGType(op.OpType),
		Payload:  op,
	}
	err := m.flood(&msg, clientsToSend)
	if err != nil {
		log.Printf("error in flooding: %s", err.Error())
	}
}

//TODO: FIX ERROR HANDLING ON MULTIPLE CLIENTS. FFS
func (m *Miner) flood(msg *tcp.Msg, peers []*tcp.Client) error {
	errors := []string{}
	//mutex := sync.Mutex{}
	//wg := sync.WaitGroup{}
	for _, client := range peers {
		//wg.Add(1)
		go func(c *tcp.Client) {
			log.Printf("flood to peer: %s ", c.TargetID)
			res := c.Send(msg)
			log.Printf("got res from peer: %s, %s ", c.TargetID, res.MSGType.String())
			// if res.MSGType == tcp.Error {
			// 	mutex.Lock()
			// 	defer mutex.Unlock()
			// 	err := res.Payload.(string)
			// 	errors = append(errors, c.TargetID+": "+err)
			// 	log.Printf("done flooding to peer: %s ", c.TargetID)
			// 	wg.Done()
			// }
		}(client)
	}
	//wg.Wait()
	return fmt.Errorf(strings.Join(errors, ", "))
}

func getRecord(i interface{}) *rfslib.Record {
	arr := i.([]interface{})
	r := rfslib.Record{}
	for i, v := range arr {
		r[i] = byte(v.(uint8))
	}
	return &r
}
func getUnusedFilenameToStore(baseFilename string) string {
	_, err := os.Stat(baseFilename)
	if os.IsNotExist(err) {
		return baseFilename
	}
	split := strings.Split(baseFilename, ".")
	iStr := split[len(split)-1]
	i, err := strconv.Atoi(iStr)
	var newFilename string
	if err != nil {
		i = 1
		newFilename = strings.Join(split, ".") + strconv.Itoa(i)
	}
	newFilename = strings.Join(split[:len(split)-1], ".") + strconv.Itoa(i)
	return getUnusedFilenameToStore(newFilename)
}
