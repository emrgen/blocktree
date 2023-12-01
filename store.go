package blocktree

import "errors"

var (
	ErrBlockNotFound = errors.New("block not found in store")
)

type StoreChange struct {
	blockChange   *BlockChange
	jsonDocChange []*JsonDocChange
	txChange      []*Transaction
}

type BlockStore interface {
	CreateSpace(space *Space) error
	CreateBlock(spaceID *SpaceID, block *Block) error
	GetBlock(spaceID *SpaceID, id BlockID) (*Block, error)
	GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error)
	GetWithFirstChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetWithLastChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetParentWithNextBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetParentWithPrevBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetAncestorEdges(spaceID *SpaceID, id []BlockID) ([]BlockEdge, error)
}

type TransactionStore interface {
	GetTransaction(spaceID *SpaceID, id *TransactionID) (*Transaction, error)
}

type JsonDocStore interface {
}

type Store interface {
	BlockStore
	TransactionStore
	JsonDocStore
	ApplyChange(space *SpaceID, change *StoreChange) error
}
