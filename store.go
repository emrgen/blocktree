package blocktree

import "errors"

var (
	ErrBlockNotFound = errors.New("block not found in store")
)

type StoreChange struct {
	blockChange   *BlockChange
	jsonDocChange []*JsonPatch
	txChange      []*Transaction
}

type BlockStore interface {
	CreateSpace(space *Space) error
	GetBlockSpaceID(id *BlockID) (*SpaceID, error)
	CreateBlock(spaceID *SpaceID, block *Block) error
	GetBlock(spaceID *SpaceID, id BlockID) (*Block, error)
	GetChildrenBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetDescendantBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetParentBlock(spaceID *SpaceID, id BlockID) (*Block, error)
	GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error)
	GetWithFirstChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetWithLastChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetParentWithNextBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetParentWithPrevBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	GetAncestorEdges(spaceID *SpaceID, id []BlockID) ([]BlockEdge, error)
}

type TransactionStore interface {
	GetLatestTransaction(spaceID *SpaceID) (*Transaction, error)
	//GetTransactions(spaceID *SpaceID, id *TransactionID) (*Transaction, error)
	PutTransactions(spaceID *SpaceID, tx []*Transaction) error
}

type JsonDocStore interface {
}

type Store interface {
	BlockStore
	TransactionStore
	JsonDocStore

	// Apply applies blocktree change to db in one transaction
	Apply(space *SpaceID, change *StoreChange) error
}
