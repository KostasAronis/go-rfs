package blockchain

import "errors"

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
	GenesisNode *Block
	Blocks      []*Block
	OpDiff      int
	NoopDiff    int
}

func (b *BlockTree) GetLongestChain() {

}
func (b *BlockTree) GetLastBlock() *Block {
	return nil
}

func (b *BlockTree) AppendBlock(block *Block) error {
	valid, err := b.ValidNode(block)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid block")
	}
	b.Blocks = append(b.Blocks, block)
	return nil
}

func (b *BlockTree) getBlockByHash(hash string) *Block {
	for _, block := range b.Blocks {
		if block.GetComputedHash() == hash {
			return block
		}
	}
	return nil
}

func (b *BlockTree) ValidNode(block *Block) (bool, error) {
	var diff int
	if block.IsOp {
		diff = b.OpDiff
	} else {
		diff = b.NoopDiff
	}
	isValid, err := block.ValidHash(diff)
	if err != nil {
		return false, err
	}
	if isValid {
		if block.PrevHash != "" {
			prevBlock := b.getBlockByHash(block.PrevHash)
			if prevBlock.IsOp {
				diff = b.OpDiff
			} else {
				diff = b.NoopDiff
			}
			return prevBlock.ValidHash(diff)
		}
	}
	return false, nil
}
