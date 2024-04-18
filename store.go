package blocktree

import (
	"errors"
)

var (
	ErrBlockNotFound = errors.New("block not found in store")
)

type storeChange struct {
	blockChange   *blockChange
	jsonDocChange []*JsonPatch
	txChange      []*Transaction
}

func (sc *storeChange) intoSyncBlocks() *SyncBlocks {
	sb := NewSyncBlocks()
	sb.children.Extend(sc.blockChange.children.ToSlice())

	for _, child := range sc.blockChange.updated.ToSlice() {
		sb.updated.Add(child.ID)
	}

	for _, child := range sc.blockChange.propSet.ToSlice() {
		sb.props.Add(child.ID)
	}

	return sb
}

type SyncBlocks struct {
	children *Set[BlockID]
	patched  *Set[BlockID]
	updated  *Set[BlockID]
	props    *Set[BlockID]
}

func NewSyncBlocks() *SyncBlocks {
	return &SyncBlocks{
		children: NewSet[BlockID](),
		patched:  NewSet[BlockID](),
		updated:  NewSet[BlockID](),
		props:    NewSet[BlockID](),
	}
}

func (sb *SyncBlocks) extend(other *SyncBlocks) {
	sb.children.Extend(other.children.ToSlice())
	sb.patched.Extend(other.patched.ToSlice())
	sb.updated.Extend(other.updated.ToSlice())
	sb.props.Extend(other.props.ToSlice())
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
	GetAncestorEdges(spaceID *SpaceID, id []BlockID) ([]blockEdge, error)
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
	Apply(space *SpaceID, change *storeChange) error
}
