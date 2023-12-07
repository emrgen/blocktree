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
	tx       *Transaction
}

func newSpaceStore() *spaceStore {
	return &spaceStore{
		children: make(map[ParentID]*btree.BTreeG[*Block]),
		blocks:   make(map[BlockID]*Block),
		parents:  make(map[BlockID]ParentID),
		props:    make(map[BlockID]map[string]interface{}),
	}
}

func (ss *spaceStore) AddBlock(block *Block) {
	ss.blocks[block.ID] = block.Clone()
	ss.parents[block.ID] = block.ParentID
	children, ok := ss.children[block.ParentID]
	if !ok {
		children = btree.NewG(10, blockLessFunc)
		ss.children[block.ParentID] = children
	}
	children.ReplaceOrInsert(block)

	if block.Props != nil {
		ss.props[block.ID] = block.Props
	}
}

func (ss *spaceStore) RemoveBlock(id BlockID) {
	block, ok := ss.blocks[id]
	if !ok {
		return
	}

	//logrus.Info("removing block from store", block.ParentID)
	delete(ss.parents, id)
	children, ok := ss.children[block.ParentID]
	if !ok {
		return
	}

	//logrus.Infof("removing block %v from parent %v", id, *block.ParentID)
	children.Delete(block)

	delete(ss.blocks, id)
}

// MemStore is a blocktree store that stores everything in memory.
type MemStore struct {
	spaces map[SpaceID]*spaceStore
}

func (ms *MemStore) GetLatestTransaction(spaceID *SpaceID) (*Transaction, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	return space.tx, nil
}

func (ms *MemStore) GetChildrenBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, err := ms.getSpace(spaceID)
	if err != nil {
		return nil, err
	}

	blocks := make([]*Block, 0)
	children, ok := space.children[id]
	if !ok {
		return blocks, nil
	}

	children.Ascend(func(item *Block) bool {
		if !item.Linked {
			blocks = append(blocks, item.Clone())
		}
		return true
	})

	return blocks, nil
}

func (ms *MemStore) GetDescendantBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, err := ms.getSpace(spaceID)
	if err != nil {
		return nil, err
	}

	blocks := make([]*Block, 0)
	ms.getDescendantBlocks(space, id, &blocks)

	return blocks, nil
}

func (ms *MemStore) getDescendantBlocks(space *spaceStore, id BlockID, blocks *[]*Block) {
	if block, ok := space.blocks[id]; ok && block != nil {
		*blocks = append(*blocks, block.Clone())
		// stop at page block, no need to go further
		if block.Type == "page" {
			return
		}
	} else {
		return
	}

	children, ok := space.children[id]
	if !ok {
		return
	}

	children.Ascend(func(item *Block) bool {
		ms.getDescendantBlocks(space, item.ID, blocks)
		return true
	})
}

func (ms *MemStore) GetParentBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	parentID, ok := space.parents[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("block %v not found", id))
	}

	return space.blocks[parentID].Clone(), nil
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
	blocks := []*Block{block.Clone()}

	children, ok := space.children[id]
	if !ok {
		return blocks, nil
	}

	if children.Len() == 0 {
		return blocks, nil
	}

	children.Ascend(func(item *Block) bool {
		blocks = append(blocks, item.Clone())
		return false
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
	blocks := []*Block{block.Clone()}

	children, ok := space.children[id]
	if !ok {
		return blocks, nil
	}

	if children.Len() == 0 {
		return blocks, nil
	}

	children.Descend(func(item *Block) bool {
		blocks = append(blocks, item.Clone())
		return false
	})

	return blocks, nil
}

func (ms *MemStore) GetParentWithNextBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, err := ms.getSpace(spaceID)
	if err != nil {
		return nil, err
	}

	blocks := make([]*Block, 0)
	parent, ok := space.parents[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("parent block not found for: %v", id))
	}
	blocks = append(blocks, space.blocks[parent].Clone())

	children, ok := space.children[parent]
	if !ok {
		return nil, errors.New(fmt.Sprintf("block siblings not found id: %v", id))
	}

	children.DescendLessOrEqual(space.blocks[id], func(item *Block) bool {
		blocks = append(blocks, item.Clone())
		if len(blocks) >= 2 {
			return false
		}
		return true
	})

	if len(blocks) == 1 {
		return nil, errors.New(fmt.Sprintf("block not found id: %v", id))
	}

	return blocks, nil
}

func (ms *MemStore) GetParentWithPrevBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, err := ms.getSpace(spaceID)
	if err != nil {
		return nil, err
	}

	blocks := make([]*Block, 0)
	parent, ok := space.parents[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("parent block not found for: %v", id))
	}
	blocks = append(blocks, space.blocks[parent].Clone())

	children, ok := space.children[parent]
	if !ok {
		return nil, errors.New(fmt.Sprintf("block siblings not found id: %v", id))
	}

	children.AscendGreaterOrEqual(space.blocks[id], func(item *Block) bool {
		blocks = append(blocks, item.Clone())
		if len(blocks) >= 2 {
			return false
		}
		return true
	})

	if len(blocks) == 1 {
		return nil, errors.New(fmt.Sprintf("block not found id: %v", id))
	}

	return blocks, nil
}

func (ms *MemStore) CreateSpace(space *Space) error {
	if _, ok := ms.spaces[space.ID]; ok {
		return errors.New(fmt.Sprintf("space %v already exists", space.ID))
	}

	ms.spaces[space.ID] = newSpaceStore()
	spaceBlock := NewBlock(space.ID, RootBlockID, "space")
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

func (ms *MemStore) Apply(spaceID *SpaceID, change *StoreChange) error {
	if change == nil {
		return errors.New("cannot apply nil change to store")
	}

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

		for _, block := range blockChange.updated.ToSlice() {
			//logrus.Infof("updating block %v", block)
			storeBlock, ok := space.blocks[block.ID]
			if !ok {
				return errors.New(fmt.Sprintf("move block not found, %v", block.ID))
			}
			space.RemoveBlock(block.ID)

			storeBlock.ParentID = block.ParentID
			storeBlock.Index = block.Index
			space.AddBlock(storeBlock)
		}
	}

	if change.txChange != nil {
		//TODO implement me
		panic("implement me")
	}

	if change.jsonDocChange != nil {
		//TODO implement me
		panic("implement me")
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

	space.AddBlock(block)

	return nil
}

func (ms *MemStore) GetBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	return space.blocks[id].Clone(), nil
}

func (ms *MemStore) GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	blocks := make([]*Block, 0, len(ids))
	for _, id := range ids {
		if block, ok := space.blocks[id]; ok {
			blocks = append(blocks, block.Clone())
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

func (ms *MemStore) getSpace(spaceID *SpaceID) (*spaceStore, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space not found: %v", *spaceID))
	}
	return space, nil
}

func (ms *MemStore) GetTransaction(spaceID *SpaceID, id *TransactionID) (*Transaction, error) {
	return nil, nil
}

func (ms *MemStore) PutTransactions(spaceID *SpaceID, txs []*Transaction) error {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		ms.spaces[*spaceID] = space
	}

	if len(txs) == 0 {
		return errors.New("transactions are empty")
	}

	space.tx = txs[len(txs)-1]
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
