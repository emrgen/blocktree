package blocktree

import (
	"errors"
	"fmt"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

var (
	_ Store = (*MemStore)(nil)
)

type spaceStore struct {
	children map[ParentID]*btree.BTreeG[*Block]
	blocks   map[BlockID]*Block
	parents  map[BlockID]ParentID
	props    map[BlockID]map[string]interface{}
	txs      map[TransactionID]*Transaction
}

func newSpaceStore() *spaceStore {
	return &spaceStore{
		children: make(map[ParentID]*btree.BTreeG[*Block]),
		blocks:   make(map[BlockID]*Block),
		parents:  make(map[BlockID]ParentID),
		props:    make(map[BlockID]map[string]interface{}),
	}
}

type MemStore struct {
	spaces map[SpaceID]*spaceStore
}

func (ms *MemStore) GetWithFirstChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	block, ok := space.blocks[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("block %v not found", id))
	}
	blocks := []*Block{block}

	children, ok := space.children[id]
	if !ok {
		return blocks, nil
	}

	if children.Len() == 0 {
		return blocks, nil
	}

	children.Ascend(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks, nil
}

func (ms *MemStore) GetWithLastChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	block, ok := space.blocks[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("block %v not found", id))
	}
	blocks := []*Block{block}

	children, ok := space.children[id]
	if !ok {
		return blocks, nil
	}

	if children.Len() == 0 {
		return blocks, nil
	}

	children.Descend(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks, nil
}

func (ms *MemStore) GetParentWithNextBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (ms *MemStore) GetParentWithPrevBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (ms *MemStore) CreateSpace(space *Space) error {
	if _, ok := ms.spaces[space.ID]; ok {
		return errors.New(fmt.Sprintf("space %v already exists", space.ID))
	}

	ms.spaces[space.ID] = newSpaceStore()
	spaceBlock := NewBlock(space.ID, nil, "space")
	err := ms.CreateBlock(&space.ID, spaceBlock)
	if err != nil {
		return err
	}

	return nil
}

func NewMemStore() *MemStore {
	return &MemStore{
		spaces: make(map[SpaceID]*spaceStore),
	}
}

func (ms *MemStore) ApplyChange(spaceID *SpaceID, change *StoreChange) error {
	logrus.Info("applying change to store")
	// all changes are part of a single transaction
	space, ok := ms.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		ms.spaces[*spaceID] = space
	}

	if change.blockChange != nil {
		blockChange := change.blockChange
		for _, block := range blockChange.inserted.ToSlice() {
			err := ms.CreateBlock(spaceID, block)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ms *MemStore) PutSpace(spaceID *SpaceID) error {
	ms.spaces[*spaceID] = newSpaceStore()
	return nil
}

func (ms *MemStore) CreateBlock(spaceID *SpaceID, block *Block) error {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		ms.spaces[*spaceID] = space
	}

	space.blocks[block.ID] = block
	if block.ParentID != nil {
		space.parents[block.ID] = *block.ParentID
		children, ok := space.children[*block.ParentID]
		if !ok {
			children = btree.NewG(10, blockLessFunc)
			space.children[*block.ParentID] = children
		}
		children.ReplaceOrInsert(block)

	}

	if block.Props != nil {
		space.props[block.ID] = block.Props
	}

	return nil
}

func (ms *MemStore) GetBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	return space.blocks[id], nil
}

func (ms *MemStore) GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	blocks := make([]*Block, 0, len(ids))
	for _, id := range ids {
		if block, ok := space.blocks[id]; ok {
			blocks = append(blocks, block)
		}
	}
	return blocks, nil
}

func (ms *MemStore) GetAncestorEdges(spaceID *SpaceID, ids []BlockID) ([]BlockEdge, error) {
	edges := make([]BlockEdge, 0)
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	for _, id := range ids {
		curr := id
		for {
			parent, ok := space.parents[curr]
			if !ok && curr != *spaceID {
				return nil, errors.New(fmt.Sprintf("non space block %v has no parent", curr))
			}

			edges = append(edges, BlockEdge{parentID: parent, childID: curr})
			if parent == *spaceID {
				break
			}
			curr = parent
		}
	}

	return edges, nil
}

func (ms *MemStore) GetTransaction(spaceID *SpaceID, id *TransactionID) (*Transaction, error) {
	return nil, nil
}

func (ms *MemStore) PutTransaction(spaceID *SpaceID, tx *Transaction) error {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		ms.spaces[*spaceID] = space
	}

	space.txs[tx.ID] = tx
	return nil
}

func (ms *MemStore) Print(spaceID *SpaceID) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		logrus.Warnf("space %v not found", *spaceID)
		return
	}

	tree := treeprint.New()
	branch := tree.AddBranch(fmt.Sprintf("space: %v", *spaceID))
	traverseStore(space, spaceID, branch)

	fmt.Println(tree.String())
}

func traverseStore(space *spaceStore, parentID *BlockID, tree treeprint.Tree) {
	children, ok := space.children[*parentID]
	if !ok {
		return
	}

	children.Ascend(func(item *Block) bool {

		branch := tree.AddBranch(fmt.Sprintf("%v: (%v) %v", item.Type, item.Index.String(), item.ID))
		traverseStore(space, &item.ID, branch)
		return true
	})
}
