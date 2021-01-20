package blockchain

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
	GenesisNode *BlockTreeNode
	Nodes       []*BlockTreeNode
}

func (b *BlockTree) GetLongestChain() {

}
func (b *BlockTree) GetLastBlock() *Block {
	return nil
}

func (b *BlockTree) AppendBlock(block *Block) {

}
