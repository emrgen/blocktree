package blocktree

type storeChange struct {
	blockChange   *blockChange
	jsonDocChange []*JsonPatch
	tx            *Transaction
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

// BlockStore is a store for blocks
type BlockStore interface {
	// CreateSpace creates a new space in the store
	CreateSpace(space *Space) error
	// GetBlockSpaceID returns the space id of the block
	GetBlockSpaceID(id *BlockID) (*SpaceID, error)
	//CreateBlock creates a new block in the store
	CreateBlock(spaceID *SpaceID, block *Block) error
	// GetBlock returns the block with the given id
	GetBlock(spaceID *SpaceID, id BlockID) (*Block, error)
	// GetChildrenBlocks returns the children of the block with the given id
	GetChildrenBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error)
	// GetLinkedBlocks returns the linked blocks of the block with the given id
	GetLinkedBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error)
	// GetDescendantBlocks returns the descendants of the block with the given id
	GetDescendantBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error)
	// GetParentBlock returns the parent of the block with the given id
	GetParentBlock(spaceID *SpaceID, id BlockID) (*Block, error)
	// GetBlocks returns the blocks with the given ids
	GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error)
	// GetWithFirstChildBlock returns the blocks with the given ids and properties
	GetWithFirstChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	// GetWithLastChildBlock returns the blocks with the given ids and properties
	GetWithLastChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	// GetParentWithNextBlock returns the parent with the next block
	GetParentWithNextBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	// GetParentWithPrevBlock returns the parent with the previous block
	GetParentWithPrevBlock(spaceID *SpaceID, id BlockID) ([]*Block, error)
	// GetAncestorEdges returns the ancestor edges of the block with the given id
	GetAncestorEdges(spaceID *SpaceID, id []BlockID) ([]blockEdge, error)
}

// TransactionStore is a store for transactions
type TransactionStore interface {
	GetTransaction(spaceID *SpaceID, id TransactionID) (*Transaction, error)
	// GetLatestTransaction returns the latest transaction in the store
	GetLatestTransaction(spaceID *SpaceID) (*Transaction, error)
	// PutTransaction puts transactions in the store
	PutTransaction(spaceID *SpaceID, tx *Transaction) error
}

// JsonDocStore is a store for JSON documents
type JsonDocStore interface {
}

type Store interface {
	BlockStore
	TransactionStore
	JsonDocStore

	// Apply applies blocktree change to db in one transaction
	Apply(tx *Transaction, change *storeChange) error
}
