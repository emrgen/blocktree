package blocktree

type storeChange struct {
	blockChange   *blockChange
	jsonDocChange []*JsonPatch
	tx            *Transaction
}

// intoSyncBlocks converts the store change into a SyncBlocks object
func (sc *storeChange) intoSyncBlocks() *SyncBlocks {
	sb := NewSyncBlocks()
	sb.children.Extend(sc.blockChange.children.ToSlice())

	//for _, child := range sc.blockChange.updated.ToSlice() {
	//	sb.updated.Add(child.ID)
	//}

	for _, child := range sc.blockChange.propSet.ToSlice() {
		sb.props.Add(child.ID)
	}

	for _, child := range sc.blockChange.inserted.ToSlice() {
		sb.inserted.Add(child.ID)
	}

	for _, child := range sc.blockChange.patched.ToSlice() {
		sb.patched.Add(child.ID)
	}

	return sb
}

// SyncBlocks is a set of blocks that have been updated and need to be synced with the clients
type SyncBlocks struct {
	children *Set[BlockID]
	inserted *Set[BlockID]
	patched  *Set[BlockID]
	updated  *Set[BlockID]
	props    *Set[BlockID]
}

// NewSyncBlocks creates a new SyncBlocks object
func NewSyncBlocks() *SyncBlocks {
	return &SyncBlocks{
		children: NewSet[BlockID](),
		inserted: NewSet[BlockID](),
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
	sb.inserted.Extend(other.inserted.ToSlice())
}

func (sb *SyncBlocks) dirty() *Set[BlockID] {
	dirty := NewSet[BlockID]()

	dirty.Extend(sb.inserted.ToSlice())
	dirty.Extend(sb.patched.ToSlice())
	dirty.Extend(sb.updated.ToSlice())
	dirty.Extend(sb.props.ToSlice())

	return dirty
}

func (sb *SyncBlocks) IsEmpty() bool {
	return sb.children.Size() == 0 && sb.patched.Size() == 0 && sb.updated.Size() == 0 && sb.props.Size() == 0
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
	// GetChildrenBlockIDs returns the children block ids of the block with the given id
	GetChildrenBlockIDs(spaceID *SpaceID, id BlockID) ([]BlockID, error)
	// GetLinkedBlocks returns the linked blocks of the block with the given id
	GetLinkedBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error)
	//GetBackLinks return the blocks the current block is linked at
	GetBackLinks(spaceID *SpaceID, id BlockID) ([]*Block, error)
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
	// GetNextTransactions returns the next transactions in the store
	GetNextTransactions(spaceID *SpaceID, id TransactionID, start, limit int) ([]*Transaction, error)
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
