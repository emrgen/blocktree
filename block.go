package blocktree

import "github.com/google/uuid"

type Block struct {
	Type     string
	ID       uuid.UUID
	ParentID uuid.UUID
	Index    *FracIndex
	Children []*Block
}

func NewBlock() *Block {
	return &Block{}
}

type blockTree struct {
	Root *blockEntry
}

func newBlockTree() *blockTree {
	return &blockTree{}
}

func (b *blockTree) addBlock(block *blockEntry) error {
	return nil
}

type blockEntry struct {
	Type       string
	ID         uuid.UUID
	ParentID   uuid.UUID
	Index      *FracIndex
	Properties map[string]interface{}
}
