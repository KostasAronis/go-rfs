package blockchainfs

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/KostasAronis/go-rfs/blockchain"
	"github.com/KostasAronis/go-rfs/filesystem"
	"github.com/KostasAronis/go-rfs/minerconfig"
	"github.com/KostasAronis/go-rfs/serialization"
)

const maxNonce uint32 = 4294967295
const goRoutineCount = 5

type BlockchainFS struct {
	//on init
	config         *minerconfig.Config
	pauseNoopChan  chan bool
	resumeNoopChan chan bool
	blockchain     *blockchain.BlockTree
	FS             *filesystem.FileSystem
	bank           map[string]int
	//on new staging
	timer              *time.Timer
	stagingOps         []*blockchain.OpRecord
	stagingFS          *filesystem.FileSystem
	stagingBank        map[string]int
	waitingClientBlock []chan *blockchain.Block
	//unused?
	producedBlock    chan *blockchain.Block
	waitingClientErr []chan error
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
	go b.mineForever(b.pauseNoopChan, b.resumeNoopChan)
	return nil
}

//AddBlock adds mined block
func (b *BlockchainFS) AddBlock(block *blockchain.Block) error {
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
		M:           &sync.Mutex{},
		GenesisNode: &genesisBlock,
		Blocks: []*blockchain.Block{
			&genesisBlock,
		},
		NoopDiff: b.config.CommonMinerConfig.PowPerNoOpBlock,
		OpDiff:   b.config.CommonMinerConfig.PowPerOpBlock,
	}
	return nil
}
func (b *BlockchainFS) Store(filename string) error {
	bytes, err := serialization.EncodeToBytes(b.blockchain)
	if err != nil {
		return err
	}
	err = serialization.WriteToFile(bytes, filename)
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
func (b *BlockchainFS) mineNOOP(block *blockchain.Block, done chan *blockchain.Block) {
	block = b.mine(block)
	done <- block
}

func (b *BlockchainFS) startMiningNoop(done chan *blockchain.Block, stop chan bool) {
	prevHash, err := b.blockchain.GetLastBlock().ComputeHash()
	if err != nil {
		panic(err)
	}
	noop := blockchain.Block{
		PrevHash: prevHash,
		MinerID:  b.config.MinerID,
		IsOp:     false,
	}
	noopMined := make(chan *blockchain.Block)
	go b.mineNOOP(&noop, noopMined)
	for {
		select {
		case <-stop:
			return
		case block := <-noopMined:
			done <- block
		}
	}
}

func canStartNoOp() bool {
	m := sync.Mutex{}
	m.Lock()
	m.Unlock()
	return true
}

func (b *BlockchainFS) mineForever(pause chan bool, resume chan bool) {
	log.Println("started mining")
	for {
		log.Println("new mining loop")
		noopMined := make(chan *blockchain.Block, 1)
		stopChan := make(chan bool, 1)
		go b.startMiningNoop(noopMined, stopChan)
		for {
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
			case <-pause:
				log.Println("noop mining paused")
				<-resume
				log.Println("noop mining resumed")
				break
			}
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

func (b *BlockchainFS) startTimer() {
	log.Println("started timer on new op batch")
	b.timer = time.NewTimer(time.Duration(b.config.CommonMinerConfig.GenOpBlockTimeout) * time.Millisecond)
	go func() {
		<-b.timer.C
		log.Println("started mining new op block")
		b.pauseNoopChan <- true
		newBlock := b.createStageBlock()
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

func (b *BlockchainFS) tryAddBlock(block *blockchain.Block) error {
	err := b.AddBlock(block)
	if err != nil {
		panic(err)
	}
	coins := b.config.CommonMinerConfig.MinedCoinsPerNoOpBlock
	if block.IsOp {
		coins = b.config.CommonMinerConfig.MinedCoinsPerOpBlock
	}
	b.bank[block.MinerID] = b.bank[block.MinerID] + coins
	return nil
}

func (b *BlockchainFS) createStageBlock() *blockchain.Block {
	b.timer = nil
	b.bank = b.stagingBank
	b.FS = b.stagingFS
	// b.stagingBank = nil
	// b.FS = nil
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
		Ops:           b.stagingOps,
		Confirmations: 0,
	}
	newBlock = *b.mine(&newBlock)
	return &newBlock
}

//DONT LOOK FURTHER (FOR NOW)
//Mine Works on solving a block nonce
func (b *BlockchainFS) mine(block *blockchain.Block) *blockchain.Block {
	var difficulty int
	if block.IsOp {
		difficulty = b.config.CommonMinerConfig.PowPerOpBlock
	} else {
		difficulty = b.config.CommonMinerConfig.PowPerNoOpBlock
	}
	block = calculateNonceParallel(block, difficulty)
	return block
}

func calculateNonceParallel(block *blockchain.Block, difficulty int) *blockchain.Block {
	out := make(chan blockchain.Block)
	for i := 0; i < goRoutineCount; i++ {
		go tryFindNonce(uint32(i), difficulty, out, *block)
	}
	minedBlock := <-out
	return &minedBlock
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
