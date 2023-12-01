package blocktree

import (
	"errors"
	"fmt"
	"github.com/google/btree"
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

func (s *MemStore) GetWithFirstChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, ok := s.spaces[*spaceID]
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

func (s *MemStore) GetWithLastChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, ok := s.spaces[*spaceID]
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

func (s *MemStore) GetParentWithNextBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (s *MemStore) GetParentWithPrevBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (s *MemStore) CreateSpace(space *Space) error {
	if _, ok := s.spaces[space.ID]; ok {
		return errors.New(fmt.Sprintf("space %v already exists", space.ID))
	}

	s.spaces[space.ID] = newSpaceStore()
	spaceBlock := NewBlock(space.ID, nil, "space")
	err := s.CreateBlock(&space.ID, spaceBlock)
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

func (s *MemStore) ApplyChange(spaceID *SpaceID, change *StoreChange) error {
	// all changes are part of a single transaction
	space, ok := s.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		s.spaces[*spaceID] = space
	}

	if change.blockChange != nil {
		change.blockChange.inserted.ForEach(func(item *Block) bool {
			return false
		})
	}

	return nil
}

func (s *MemStore) PutSpace(spaceID *SpaceID) error {
	s.spaces[*spaceID] = newSpaceStore()
	return nil
}

func (s *MemStore) CreateBlock(spaceID *SpaceID, block *Block) error {
	space, ok := s.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		s.spaces[*spaceID] = space
	}

	space.blocks[block.ID] = block
	if block.ParentID != nil {
		space.parents[block.ID] = *block.ParentID
		children, ok := space.children[*block.ParentID]
		if !ok {
			children = btree.NewG(2, blockLessFunc)
			space.children[*block.ParentID] = children
		}
		children.ReplaceOrInsert(block)

	}

	if block.Props != nil {
		space.props[block.ID] = block.Props
	}

	return nil
}

func (s *MemStore) GetBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	space, ok := s.spaces[*spaceID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("space %v not found", *spaceID))
	}

	return space.blocks[id], nil
}

func (s *MemStore) GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error) {
	space, ok := s.spaces[*spaceID]
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

func (s *MemStore) GetAncestorEdges(spaceID *SpaceID, ids []BlockID) ([]BlockEdge, error) {
	edges := make([]BlockEdge, 0)
	space, ok := s.spaces[*spaceID]
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

func (s *MemStore) GetTransaction(spaceID *SpaceID, id *TransactionID) (*Transaction, error) {
	return nil, nil
}

func (s *MemStore) PutTransaction(spaceID *SpaceID, tx *Transaction) error {
	space, ok := s.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		s.spaces[*spaceID] = space
	}

	space.txs[tx.ID] = tx
	return nil
}
