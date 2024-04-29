package blocktree

import (
	"errors"
	"sort"
)

type Api struct {
	store Store
}

func NewApi(store Store) *Api {
	return &Api{
		store: store,
	}
}

// Apply applies the given transactions to the store.
func (a *Api) Apply(transactions ...*Transaction) (*SyncBlocks, error) {
	sb := SyncBlocks{
		children: NewSet[BlockID](),
		patched:  NewSet[BlockID](),
		updated:  NewSet[BlockID](),
		props:    NewSet[BlockID](),
	}

	for _, tx := range transactions {
		change, err := tx.prepare(a.store)
		if err != nil {
			if errors.Is(err, ErrDetectedCycle) || errors.Is(err, ErrCreatesCycle) {
				continue
			}

			return nil, err
		}

		err = a.store.Apply(tx, change)
		if err != nil {
			return nil, err
		}

		sb.extend(change.intoSyncBlocks())
	}

	return &sb, nil
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

// GetDescendantBlocks returns the descendant blocks of the block with the given ID.
// The descendant blocks are the children blocks, the children of the children blocks, and so on.
func (a *Api) GetDescendantBlocks(spaceID, blockID BlockID) ([]*Block, error) {
	return a.store.GetDescendantBlocks(&spaceID, blockID)
}

type BlockUpdates struct {
	Children map[BlockID][]BlockID
	Blocks   map[BlockID]*Block
}

func (a *Api) GetUpdates(spaceID SpaceID, txID TransactionID) (*BlockUpdates, error) {
	panic("not implemented")
}
