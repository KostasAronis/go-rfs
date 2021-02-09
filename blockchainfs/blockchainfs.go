// Package blockchainfs
package blockchainfs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
	"github.com/KostasAronis/go-rfs/blockchain"
	"github.com/KostasAronis/go-rfs/filesystem"
	"github.com/KostasAronis/go-rfs/minerconfig"
	"github.com/KostasAronis/go-rfs/serialization"
)

const maxNonce uint32 = 4294967295
const goRoutineCount = 2

//BlockchainFS contains most of the logic for the operations between the blockchain and the fs
//TODO: nil checks on type properties is communicating by sharing memory. It should be the oposite. NEEDS MOAR CHAN!!!
//TODO: less type properties, more functional methods! Do not keep channels _loose_ on the _global_ scope.
type BlockchainFS struct {
	BlockToFlood chan *blockchain.Block
	GovecLogger  *govec.GoLog
	//on init
	config          *minerconfig.Config
	pauseNoopChan   chan bool
	resumeNoopChan  chan bool
	resetOpMineChan chan bool
	isMiningOp      uint32
	blockchain      *blockchain.BlockTree
	FS              *filesystem.FileSystem
	bank            map[string]int
	//on new staging
	timer              *time.Timer
	stagingOps         []*blockchain.OpRecord
	stagingFS          *filesystem.FileSystem
	stagingBank        map[string]int
	waitingClientBlock []chan *blockchain.Block
	//unused?
	currentlyMinedHash atomic.Value
}

//Init initializes BlockchainFS
func (b *BlockchainFS) Init(config *minerconfig.Config) error {
	b.config = config
	b.FS = &filesystem.FileSystem{}
	b.FS.Init()
	err := b.initBlockchain()
	if err != nil {
		return err
	}
	b.bank = map[string]int{}
	b.pauseNoopChan = make(chan bool)
	b.resumeNoopChan = make(chan bool)
	b.resetOpMineChan = make(chan bool)
	atomic.StoreUint32(&b.isMiningOp, 0)
	go b.mineForever()
	go func() {
		SIGINT := make(chan os.Signal, 1)
		signal.Notify(SIGINT, os.Interrupt)
		<-SIGINT
		err := b.Store("./logs/data" + b.config.MinerID + ".bin")
		if err != nil {
			panic(err)
		}
		os.Exit(42)
	}()
	return nil
}

//AddBlock adds mined block
func (b *BlockchainFS) addBlock(block *blockchain.Block) error {
	err := b.blockchain.AppendBlock(block)
	if err != nil {
		return err
	}
	return nil
}

func (b *BlockchainFS) TryStageOp(op *blockchain.OpRecord) (chan *blockchain.Block, error) {
	if b.timer == nil {
		b.initStaging()
	}
	if op.OpType == blockchain.CreateFile {
		coins, ok := b.stagingBank[op.MinerID]
		if !ok {
			return nil, errors.New(op.MinerID + " invalid coin count")
		}
		if coins-b.config.CommonMinerConfig.NumCoinsPerFileCreate < 0 {
			return nil, errors.New(op.MinerID + " invalid coin count")
		}
		stagingCopy := b.stagingFS.Clone()
		_, err := b.stagingFS.AddFile(op.Filename)
		if err != nil {
			b.stagingFS = stagingCopy
			return nil, err
		}
		b.stagingBank[op.MinerID] = coins - b.config.CommonMinerConfig.NumCoinsPerFileCreate
	}
	if op.OpType == blockchain.AppendRec {
		stagingCopy := b.stagingFS.Clone()
		_, err := b.stagingFS.AppendRecord(op.Filename, op.Record)
		if err != nil {
			b.stagingFS = stagingCopy
			return nil, err
		}
	}
	if b.timer == nil {
		b.startTimer()
	}
	b.stagingOps = append(b.stagingOps, op)
	blockChan := make(chan *blockchain.Block)
	b.waitingClientBlock = append(b.waitingClientBlock, blockChan)
	return blockChan, nil
}
func (b *BlockchainFS) initBlockchain() error {
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
	if b.config.CommonMinerConfig.GenesisBlockHash != genesisBlockHash {
		return fmt.Errorf("incorrect genesis block hash given: %s != %s :computed", b.config.CommonMinerConfig.GenesisBlockHash, genesisBlockHash)
	}
	b.blockchain = &blockchain.BlockTree{
		GenesisNode: &genesisBlock,
		Blocks: []*blockchain.Block{
			&genesisBlock,
		},
		NoopDiff: b.config.CommonMinerConfig.PowPerNoOpBlock,
		OpDiff:   b.config.CommonMinerConfig.PowPerOpBlock,
	}
	b.blockchain.Init()
	return nil
}

func (b *BlockchainFS) BlockExists(hash string) bool {
	block := b.blockchain.GetBlockByHash(hash)
	return block != nil
}

func (b *BlockchainFS) Store(filename string) error {
	bytes, err := serialization.EncodeToBytes(b.blockchain)
	if err != nil {
		return err
	}
	err = serialization.WriteToFile(bytes, filename)
	log.Println("stored")
	return err
}
func (b *BlockchainFS) Restore(filename string) error {
	bytes, err := serialization.ReadFromFile(filename)
	if err != nil {
		return err
	}
	blockTree, err := serialization.DecodeToBlockTree(bytes)
	if err != nil {
		return err
	}
	b.blockchain = blockTree
	return nil
}

func (b *BlockchainFS) startTimer() {
	log.Println("started timer on new op batch")
	b.timer = time.NewTimer(time.Duration(b.config.CommonMinerConfig.GenOpBlockTimeout) * time.Millisecond)
	go func() {
		<-b.timer.C
		log.Println("started mining new op block")
		b.pauseNoopChan <- true
		b.timer = nil
		b.bank = b.stagingBank
		b.FS = b.stagingFS
		newBlock := b.createStageBlock(b.stagingOps)
		hash, err := newBlock.ComputeHash()
		if err != nil {
			panic(err)
		}
		err = b.tryAddBlock(newBlock)
		if err != nil {
			panic(err)
		}
		log.Println("added op block")
		b.resumeNoopChan <- true
		log.Printf("mined op block %s\n", hash)
		for _, c := range b.waitingClientBlock {
			c <- newBlock
		}
	}()
}

func (b *BlockchainFS) createStageBlock(stagingOps []*blockchain.OpRecord) *blockchain.Block {
	atomic.StoreUint32(&b.isMiningOp, 1)
	prevBlock := b.blockchain.GetLastBlock()
	prevHash, err := prevBlock.ComputeHash()
	if err != nil {
		panic(err)
	}
	newBlock := blockchain.Block{
		PrevHash:      prevHash,
		MinerID:       b.config.MinerID,
		Nonce:         0,
		IsOp:          true,
		Ops:           stagingOps,
		Confirmations: 0,
	}
	mined := make(chan *blockchain.Block)
	stopChan := make(chan bool)
	go b.mine(&newBlock, mined, stopChan)
	select {
	case minedBlock := <-mined:
		atomic.StoreUint32(&b.isMiningOp, 0)
		return minedBlock
	case <-b.resetOpMineChan:
		stopChan <- true
		return b.createStageBlock(stagingOps)
	}
}

func (b *BlockchainFS) startMiningNoop(out chan *blockchain.Block, stop chan bool) {
	prevHash, err := b.blockchain.GetLastBlock().ComputeHash()
	if err != nil {
		panic(err)
	}
	noop := blockchain.Block{
		PrevHash: prevHash,
		MinerID:  b.config.MinerID,
		IsOp:     false,
	}
	go b.mine(&noop, out, stop)
}

func (b *BlockchainFS) mineForever() {
	log.Println("started mining")
	for {
		log.Println("new mining loop")
		noopMined := make(chan *blockchain.Block, 1)
		stopChan := make(chan bool, 1)
		go b.startMiningNoop(noopMined, stopChan)
		for {
			log.Println("for_START")
			select {
			case noopBlock := <-noopMined:
				h, err := noopBlock.ComputeHash()
				if err != nil {
					panic(err)
				}
				log.Printf("mined noop block: %s\n", h)
				err = b.tryAddBlock(noopBlock)
				if err != nil {
					panic(err)
				}
				log.Printf("added noop block: %s\n", h)
				break
			case <-b.pauseNoopChan:
				stopChan <- true
				log.Println("noop mining paused")
				<-b.resumeNoopChan
				log.Println("noop mining resumed")
				break
			}
			log.Println("for_END")
			break
		}
	}
}

func (b *BlockchainFS) initStaging() {
	b.stagingFS = b.FS.Clone()
	b.stagingBank = map[string]int{}
	for k, v := range b.bank {
		b.stagingBank[k] = v
	}

	b.waitingClientBlock = []chan *blockchain.Block{}
}

func (b *BlockchainFS) tryAddBlock(block *blockchain.Block) error {
	err := b.addBlock(block)
	if err != nil {
		return err
	}
	coins := b.config.CommonMinerConfig.MinedCoinsPerNoOpBlock
	if block.IsOp {
		coins = b.config.CommonMinerConfig.MinedCoinsPerOpBlock
	}
	b.bank[block.MinerID] = b.bank[block.MinerID] + coins
	log.Println("flooding block")
	b.BlockToFlood <- block
	return nil
}

func (b *BlockchainFS) AddExternalBlock(block *blockchain.Block) error {
	//TODO:!!! this should be pause-start and not reset (WHAT IF Mining was faster than opchecking & block adding)...
	if atomic.LoadUint32(&b.isMiningOp) == 1 {
		log.Println("reseting op for External block")
		b.resetOpMineChan <- true
	} else {
		log.Println("pausing noop for External block")
		b.pauseNoopChan <- true
	}
	err := b.tryAddBlock(block)
	if err != nil {
		return err
	}
	if atomic.LoadUint32(&b.isMiningOp) != 1 {
		b.resumeNoopChan <- true
	}
	return nil
}

//DONT LOOK FURTHER (FOR NOW)
//Mine Works on solving a block nonce
func (b *BlockchainFS) mine(block *blockchain.Block, out chan *blockchain.Block, stop chan bool) {
	var difficulty int
	if block.IsOp {
		difficulty = b.config.CommonMinerConfig.PowPerOpBlock
	} else {
		difficulty = b.config.CommonMinerConfig.PowPerNoOpBlock
	}
	parallelOut := make(chan *blockchain.Block)
	go calculateNonceParallel(block, difficulty, parallelOut, stop)
	blockType := "NoOp"
	if block.IsOp {
		blockType = "Op"
	}
	for {
		select {
		case <-stop:
			b.GovecLogger.LogLocalEvent("stopping mining "+blockType+" block", govec.GoLogOptions{Priority: govec.INFO})
		case minedBlock := <-parallelOut:
			b.GovecLogger.LogLocalEvent("done mining "+blockType+" block", govec.GoLogOptions{Priority: govec.INFO})
			out <- minedBlock
		}
	}
}

func calculateNonceParallel(block *blockchain.Block, difficulty int, out chan *blockchain.Block, stop chan bool) {
	parallelOut := make(chan blockchain.Block)
	for i := 0; i < goRoutineCount; i++ {
		go tryFindNonce(uint32(i), difficulty, parallelOut, *block)
	}
	select {
	case <-stop:
		return
	case minedBlock := <-parallelOut:
		out <- &minedBlock
	}
}

func tryFindNonce(goID uint32, difficulty int, out chan blockchain.Block, block blockchain.Block) {
	for i := uint32(goID * (maxNonce / goRoutineCount)); i < (goID+1)*(maxNonce/goRoutineCount); i++ {
		block.Nonce = uint32(i)
		valid, err := block.HasValidNonce(difficulty)
		if err != nil {
			err := fmt.Errorf("Miner: tryFindNonce: error computing hash for block! %s", err.Error())
			log.Println(err)
			panic(err)
		}
		if valid {
			out <- block
			return
		}
	}
}

func (b *BlockchainFS) tryAddOp(op *blockchain.OpRecord) (*filesystem.FileSystem, map[string]int, error) {
	stagingFS := b.FS.Clone()
	stagingBank := map[string]int{}
	for k, v := range b.bank {
		stagingBank[k] = v
	}
	if op.OpType == blockchain.CreateFile {
		coins, ok := stagingBank[op.MinerID]
		if !ok {
			return nil, nil, errors.New(op.MinerID + " invalid coin count")
		}
		if coins-b.config.CommonMinerConfig.NumCoinsPerFileCreate < 0 {
			return nil, nil, errors.New(op.MinerID + " invalid coin count")
		}
		_, err := stagingFS.AddFile(op.Filename)
		if err != nil {
			return nil, nil, err
		}
		stagingBank[op.MinerID] = coins - b.config.CommonMinerConfig.NumCoinsPerFileCreate
	}
	if op.OpType == blockchain.AppendRec {
		_, err := stagingFS.AppendRecord(op.Filename, op.Record)
		if err != nil {
			return nil, nil, err
		}
	}
	return stagingFS, stagingBank, nil
}
