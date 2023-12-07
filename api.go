package blocktree

import (
	"errors"
	"sort"
)

type Api struct {
	store Store
}

func New(store Store) *Api {
	return &Api{
		store: store,
	}
}

func (a *Api) Apply(transactions ...*Transaction) error {
	for _, tx := range transactions {
		change, err := tx.Prepare(a.store)
		if err != nil {
			if errors.Is(err, ErrDetectedCycle) || errors.Is(err, ErrCreatesCycle) {
				continue
			}

			return err
		}
		err = a.store.Apply(&tx.SpaceID, change)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Api) CreateSpace(spaceID SpaceID, name string) error {
	return a.store.CreateSpace(&Space{
		ID:   spaceID,
		Name: name,
	})
}

func (a *Api) GetBlock(spaceID, blockID BlockID) (*Block, error) {
	return a.store.GetBlock(&spaceID, blockID)
}

func (a *Api) GetBlockSpaceID(blockID BlockID) (*SpaceID, error) {
	return a.store.GetBlockSpaceID(&blockID)
}

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

func (a *Api) GetDescendantBlocks(spaceID, blockID BlockID) ([]*Block, error) {
	return a.store.GetDescendantBlocks(&spaceID, blockID)
}
