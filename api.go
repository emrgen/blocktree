package blocktree

import (
	"errors"
	"sort"
)

type Api struct {
	store     Store
	publisher PublishSyncBlocks
}

func NewApi(store Store) *Api {
	return &Api{
		store:     store,
		publisher: NewNullPublisher(),
	}
}

func NewApiWithPublisher(store Store, publisher PublishSyncBlocks) *Api {
	return &Api{
		store:     store,
		publisher: publisher,
	}
}

// Apply applies the given transactions to the store.
func (a *Api) Apply(transactions ...*Transaction) (*SyncBlocks, error) {
	sb := NewSyncBlocks()

	for _, tx := range transactions {
		change, err := tx.prepare(a.store)
		if err != nil {
			if errors.Is(err, ErrDetectedCycle) || errors.Is(err, ErrCreatesCycle) {
				continue
			}

			err := a.publisher.Publish(sb)
			if err != nil {
				return nil, errors.Join(err, ErrFailedToPublish)
			}

			return nil, err
		}

		err = a.store.Apply(tx, change)
		if err != nil {
			return nil, err
		}

		sb.extend(change.intoSyncBlocks())
	}

	err := a.publisher.Publish(sb)
	if err != nil {
		return nil, ErrFailedToPublish
	}

	return sb, nil
}

// CreateSpace creates a new space with the given ID and name.
func (a *Api) CreateSpace(spaceID SpaceID, name string) error {
	return a.store.CreateSpace(&Space{
		ID:   spaceID,
		Name: name,
	})
}

// GetBlock returns the block with the given ID.
func (a *Api) GetBlock(spaceID, blockID BlockID) (*Block, error) {
	return a.store.GetBlock(&spaceID, blockID)
}

// GetBlocks returns the blocks with the given IDs.
func (a *Api) GetBlocks(spaceID SpaceID, blockIDs ...BlockID) ([]*Block, error) {
	return a.store.GetBlocks(&spaceID, blockIDs)
}

// GetBlockSpaceID returns the space ID of the block with the given ID.
func (a *Api) GetBlockSpaceID(blockID BlockID) (*SpaceID, error) {
	return a.store.GetBlockSpaceID(&blockID)
}

// GetChildrenBlocks returns the children blocks of the block with the given ID.
func (a *Api) GetChildrenBlocks(spaceID, blockID BlockID) ([]*Block, error) {
	blocks, err := a.store.GetChildrenBlocks(&spaceID, blockID)
	if err != nil {
		return nil, err
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Index.Compare(blocks[j].Index) < 0
	})

	return blocks, err
}

// GetLinkedBlocks returns the linked block of the block with the given ID
func (a *Api) GetLinkedBlocks(spaceID, blockID BlockID) ([]*Block, error) {
	return a.store.GetLinkedBlocks(&spaceID, blockID)
}

// GetBackLinks returns the blocks the block is linked from
func (a *Api) GetBackLinks(spaceID, blockID BlockID) ([]*Block, error) {
	return a.store.GetBackLinks(&spaceID, blockID)
}

// GetDescendantBlocks returns the descendant blocks of the block with the given ID.
// The descendant blocks are the children blocks, the children of the children blocks, and so on.
func (a *Api) GetDescendantBlocks(spaceID, blockID BlockID) ([]*Block, error) {
	return a.store.GetDescendantBlocks(&spaceID, blockID)
}

// GetUpdates returns the updates since the given transaction ID.
func (a *Api) GetUpdates(spaceID SpaceID, txID TransactionID) (*BlockUpdates, error) {
	txs := make([]*Transaction, 0)
	for i := 0; ; i++ {
		nextTxs, err := a.store.GetNextTransactions(&spaceID, txID, i, 100)
		if err != nil {
			return nil, err
		}
		if (len(nextTxs)) == 0 {
			break
		}
		txs = append(txs, nextTxs...)
	}

	updates := NewSyncBlocks()
	for _, tx := range txs {
		updates.extend(tx.changes)
	}

	parenIDs := updates.children.ToSlice()
	dirtyIDs := updates.dirty().ToSlice()

	childrenMap := make(map[BlockID][]BlockID)
	for _, parentID := range parenIDs {
		children, err := a.store.GetChildrenBlockIDs(&spaceID, parentID)
		if err != nil {
			return nil, err
		}
		childrenMap[parentID] = children
	}

	blocks, err := a.store.GetBlocks(&spaceID, dirtyIDs)
	if err != nil {
		return nil, err
	}
	blockMap := make(map[BlockID]*Block)
	for _, block := range blocks {
		blockMap[block.ID] = block
	}

	return &BlockUpdates{
		Children: childrenMap,
		Blocks:   blockMap,
	}, nil
}

type BlockUpdates struct {
	Children map[BlockID][]BlockID
	Blocks   map[BlockID]*Block
}
