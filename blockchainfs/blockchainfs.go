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
)

const maxNonce uint32 = 4294967295
const goRoutineCount = 5

type BlockchainFS struct {
	//on init
	config        *minerconfig.Config
	breakNoopChan chan bool
	blockchain    *blockchain.BlockTree
	FS            *filesystem.FileSystem
	bank          map[string]int
	//on new staging
	timer       *time.Timer
	stagingOps  []*blockchain.OpRecord
	stagingFS   *filesystem.FileSystem
	stagingBank map[string]int
	//unused?
	ProducedBlock      chan *blockchain.Block
	waitingClientErr   []chan error
	waitingClientBlock []chan *blockchain.Block
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
	b.breakNoopChan = make(chan bool)
	b.mineForever(b.breakNoopChan)
	return nil
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

func (b *BlockchainFS) mineNOOP(block *blockchain.Block, done chan *blockchain.Block) {
	block = b.mine(block)
	done <- block
}

func (b *BlockchainFS) startMiningNoop(done chan *blockchain.Block, stop chan bool) {
	noop := blockchain.Block{
		PrevHash: b.blockchain.GetLastBlock().GetComputedHash(),
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
func (b *BlockchainFS) mineForever(pause chan bool) {
	log.Println("started mining")
	for {
		log.Println("new mining loop")
		noopMined := make(chan *blockchain.Block, 1)
		stopChan := make(chan bool, 1)
		go b.startMiningNoop(noopMined, stopChan)
		for {
			select {
			case noopBlock := <-noopMined:
				h := noopBlock.GetComputedHash()
				log.Printf("mined noop block: %s", h)
				err := b.AddBlock(noopBlock)
				if err != nil {
					panic(err)
				}
				b.bank[b.config.MinerID] = b.bank[b.config.MinerID] + b.config.CommonMinerConfig.MinedCoinsPerNoOpBlock
				break
			case <-pause:
				break
			}
			break
		}
	}
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

func (b *BlockchainFS) initStaging() {
	b.stagingFS = b.FS.Clone()
	b.stagingBank = map[string]int{}
	for k, v := range b.bank {
		b.stagingBank[k] = v
	}

	b.waitingClientBlock = []chan *blockchain.Block{}
}

func (b *BlockchainFS) startTimer() {
	b.timer = time.NewTimer(time.Duration(b.config.CommonMinerConfig.GenOpBlockTimeout) * time.Millisecond)
	go func() {
		<-b.timer.C
		newBlock := b.createStageBlock()
		b.breakNoopChan <- true
		for _, c := range b.waitingClientBlock {
			c <- newBlock
		}
		b.ProducedBlock <- newBlock
		b.AddBlock(newBlock)
	}()
}

func (b *BlockchainFS) createStageBlock() *blockchain.Block {
	b.timer = nil
	b.bank = b.stagingBank
	b.FS = b.stagingFS
	b.stagingBank = nil
	b.FS = nil
	prevBlock := b.blockchain.GetLastBlock()
	newBlock := blockchain.Block{
		PrevHash:      prevBlock.GetComputedHash(),
		MinerID:       b.config.MinerID,
		Nonce:         0,
		IsOp:          true,
		Ops:           b.stagingOps,
		Confirmations: 0,
	}
	b.mine(&newBlock)
	return &newBlock
}

//AddBlock adds mined block
func (b *BlockchainFS) AddBlock(block *blockchain.Block) error {
	err := b.blockchain.AppendBlock(block)
	if err != nil {
		return err
	}
	return nil
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
