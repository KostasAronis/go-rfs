package blockchain

import (
	"errors"
	"fmt"
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
	M           *sync.Mutex
	GenesisNode *Block
	Blocks      []*Block
	OpDiff      int
	NoopDiff    int
}

func (b *BlockTree) getLongestChain() []*Block {
	maxChain := []*Block{}
	for _, chain := range b.generateChains() {
		if len(chain) > len(maxChain) {
			maxChain = chain
		}
	}
	return maxChain
}

func (b *BlockTree) GetLastBlock() *Block {
	b.M.Lock()
	defer b.M.Unlock()
	longestChain := b.getLongestChain()
	if len(longestChain) > 0 {
		return longestChain[len(longestChain)-1]
	}
	return b.GenesisNode
}

func (b *BlockTree) AppendBlock(block *Block) error {
	b.M.Lock()
	defer b.M.Unlock()
	valid, err := b.validNode(block, 1)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid block")
	}
	for _, existingBlock := range b.Blocks {
		if existingBlock.hash == block.hash {
			return nil
		}
	}
	b.Blocks = append(b.Blocks, block)
	return nil
}

func (b *BlockTree) generateChains() [][]*Block {
	chains := [][]*Block{}
	genesisChain := []*Block{b.GenesisNode}
	chains = append(chains, genesisChain)
	for _, newBlock := range b.Blocks {
		if newBlock.hash == b.GenesisNode.hash {
			continue
		}
		added := false
		for cIdx, chain := range chains {
			for bIdx, block := range chain {
				if block.hash == newBlock.PrevHash {
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

func (b *BlockTree) getBlockByHash(hash string) *Block {
	for _, block := range b.Blocks {
		if block.GetComputedHash() == hash {
			return block
		}
	}
	return nil
}

func (b *BlockTree) validNode(block *Block, backCheck int) (bool, error) {
	if block.hash == b.GenesisNode.hash {
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
				prevBlock := b.getBlockByHash(block.PrevHash)
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
