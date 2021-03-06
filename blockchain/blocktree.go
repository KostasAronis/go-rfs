package blockchain

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
)

/*
	blockchain
	TODO: inser
*/
//Blocktree
type BlockTreeNode struct {
	Block    *Block
	Children []*Block
}

type BlockTree struct {
	m           *sync.RWMutex
	GenesisNode *Block
	Blocks      []*Block
	OpDiff      int
	NoopDiff    int
}

func (b *BlockTree) Init() {
	b.m = &sync.RWMutex{}
}

func (b *BlockTree) GetBlockByHash(hash string) *Block {
	for _, block := range b.Blocks {
		h, err := block.ComputeHash()
		if err != nil {
			panic(err)
		}
		if h == hash {
			return block
		}
	}
	return nil
}

func (b *BlockTree) GetLastBlock() *Block {
	b.m.RLock()
	defer b.m.RUnlock()
	longestChain := b.getLongestChain()
	if len(longestChain) > 0 {
		return longestChain[len(longestChain)-1]
	}
	return b.GenesisNode
}

func (b *BlockTree) AppendBlock(block *Block) error {
	b.m.Lock()
	defer b.m.Unlock()
	valid, err := b.validNode(block, 1)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid block")
	}
	for _, existingBlock := range b.Blocks {
		hash, err := block.ComputeHash()
		if err != nil {
			return err
		}
		eHash, err := existingBlock.ComputeHash()
		if err != nil {
			return err
		}
		if eHash == hash {
			return nil
		}
	}
	b.Blocks = append(b.Blocks, block)
	return nil
}
func (b *BlockTree) getLongestChain() []*Block {
	maxLength := 0
	maxChain := [][]*Block{}
	for _, chain := range b.generateChains() {
		if len(chain) == maxLength {
			maxChain = append(maxChain, chain)
		}
		if len(chain) > maxLength {
			maxLength = len(chain)
			maxChain = [][]*Block{chain}
		}
	}
	return maxChain[rand.Intn(len(maxChain))]
}

func (b *BlockTree) generateChains() [][]*Block {
	chains := [][]*Block{}
	genesisChain := []*Block{b.GenesisNode}
	chains = append(chains, genesisChain)
	for _, newBlock := range b.Blocks {
		newHash, err := newBlock.ComputeHash()
		if err != nil {
			panic(err)
		}
		genHash, err := b.GenesisNode.ComputeHash()
		if err != nil {
			panic(err)
		}
		if newHash == genHash {
			continue
		}
		added := false
		for cIdx, chain := range chains {
			for bIdx, block := range chain {
				hash, err := block.ComputeHash()
				if err != nil {
					panic(err)
				}
				if hash == newBlock.PrevHash {
					if len(chain) > bIdx+1 { //create a new chain
						newChain := []*Block{}
						copy(newChain, chain)
						newChain = append(newChain, newBlock)
						chains = append(chains, newChain)
					} else {
						chains[cIdx] = append(chain, newBlock)
					}
					added = true
				}
			}
		}
		if !added {
			panic(fmt.Errorf("WTF PrevHash %s not found", newBlock.PrevHash))
		}
	}
	return chains
}

func (b *BlockTree) validNode(block *Block, backCheck int) (bool, error) {
	hash, err := block.ComputeHash()
	if err != nil {
		return false, err
	}
	gHash, err := b.GenesisNode.ComputeHash()
	if err != nil {
		return false, err
	}
	if hash == gHash {
		return true, nil
	}
	var diff int
	if block.IsOp {
		diff = b.OpDiff
	} else {
		diff = b.NoopDiff
	}
	isValid, err := block.IsValid(diff)
	if err != nil {
		return false, err
	}
	if isValid {
		if backCheck > 0 {
			if block.PrevHash != "" {
				prevBlock := b.GetBlockByHash(block.PrevHash)
				if prevBlock == nil {
					log.Printf("NIL PREVIOUS BLOCK FROM MINER: %s WITH HASH: %s WITH PREV: %s", block.MinerID, hash, block.PrevHash)
					return false, errors.New("NIL PREVIOUS BLOCK")
				}
				if prevBlock.IsOp {
					diff = b.OpDiff
				} else {
					diff = b.NoopDiff
				}
				return b.validNode(prevBlock, backCheck-1)
			}
		}
		return true, nil
	}
	return false, nil
}
