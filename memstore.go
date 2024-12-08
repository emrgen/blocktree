package blocktree

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

var (
	_ Store = (*MemStore)(nil)
)

type spaceStore struct {
	children  map[ParentID]*btree.BTreeG[*Block]
	blocks    map[BlockID]*Block
	parents   map[BlockID]ParentID
	props     map[BlockID][]byte
	backLinks map[BlockID]Set[BlockID]
	txs       []*Transaction
}

func newSpaceStore() *spaceStore {
	timestamp, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	return &spaceStore{
		children:  make(map[ParentID]*btree.BTreeG[*Block]),
		blocks:    make(map[BlockID]*Block),
		parents:   make(map[BlockID]ParentID),
		props:     make(map[BlockID][]byte),
		backLinks: make(map[BlockID]Set[BlockID]),
		txs: []*Transaction{{
			ID:      uuid.Nil,
			SpaceID: SpaceID{},
			UserID:  uuid.Nil,
			Time:    timestamp,
			Ops:     nil,
		}},
	}
}

func (ss *spaceStore) equals(other *spaceStore) bool {
	if len(ss.blocks) != len(other.blocks) {
		return false
	}

	for id, block := range ss.blocks {
		otherBlock, ok := other.blocks[id]
		if !ok {
			logrus.Debugf("blocks are different %v, %v", block, otherBlock)
			return false
		}

		if !reflect.DeepEqual(block, otherBlock) {
			return false
		}

		//if !reflect.DeepEqual(ss.props[id], other.props[id]) {
		//	return false
		//}
	}

	return true
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
		ss.props[block.ID] = block.Props.Bytes()
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
	spaces     map[SpaceID]*spaceStore
	blockSpace map[BlockID]SpaceID
}

func NewMemStore() *MemStore {
	return &MemStore{
		spaces:     make(map[SpaceID]*spaceStore),
		blockSpace: make(map[BlockID]SpaceID),
	}
}

// Equals compares two MemStore instances.
func (ms *MemStore) Equals(other *MemStore) bool {
	if len(ms.spaces) != len(other.spaces) {
		return false
	}

	for id, space := range ms.spaces {
		otherSpace, ok := other.spaces[id]
		if !ok {
			return false
		}
		if !space.equals(otherSpace) {
			return false
		}
	}

	return true
}

func (ms *MemStore) GetLatestTransaction(spaceID *SpaceID) (*Transaction, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	return space.txs[len(space.txs)-1], nil
}

func (ms *MemStore) GetBlockSpaceID(id *BlockID) (*SpaceID, error) {
	spaceID, ok := ms.blockSpace[*id]
	if !ok {
		return nil, fmt.Errorf("space id for block is not found, %v", *id)
	}

	return &spaceID, nil
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

func (ms *MemStore) GetChildrenBlockIDs(spaceID *SpaceID, id BlockID) ([]BlockID, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	children, ok := space.children[id]
	if !ok {
		return nil, fmt.Errorf("block %v not found", id)
	}

	ids := make([]BlockID, 0)
	children.Ascend(func(item *Block) bool {
		ids = append(ids, item.ID)
		return true
	})

	return ids, nil
}

func (ms *MemStore) GetLinkedBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
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
		if item.Linked {
			blocks = append(blocks, item.Clone())
		}
		return true
	})

	return blocks, nil
}

func (ms *MemStore) GetBackLinks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, err := ms.getSpace(spaceID)
	if err != nil {
		return nil, err
	}

	if s, ok := space.backLinks[id]; !ok {
		return []*Block{}, nil
	} else {

		blocks, err := ms.GetBlocks(spaceID, s.ToSlice())
		if err != nil {
			return nil, err
		}

		return blocks, nil
	}
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

	} else {
		return
	}

	children, ok := space.children[id]
	if !ok {
		return
	}

	children.Ascend(func(item *Block) bool {
		// stop at page block, no need to go further
		if item.Type == "page" {
			*blocks = append(*blocks, item.Clone())
			return true
		}

		ms.getDescendantBlocks(space, item.ID, blocks)
		return true
	})
}

func (ms *MemStore) GetParentBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	parentID, ok := space.parents[id]
	if !ok {
		return nil, fmt.Errorf("block %v not found", id)
	}

	return space.blocks[parentID].Clone(), nil
}

func (ms *MemStore) GetWithFirstChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	block, ok := space.blocks[id]
	if !ok {
		return nil, fmt.Errorf("block %v not found", id)
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
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	block, ok := space.blocks[id]
	if !ok {
		return nil, fmt.Errorf("block %v not found", id)
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
		return nil, fmt.Errorf("parent block not found for: %v", id)
	}
	blocks = append(blocks, space.blocks[parent].Clone())

	children, ok := space.children[parent]
	if !ok {
		return nil, fmt.Errorf("block siblings not found id: %v", id)
	}

	children.AscendGreaterOrEqual(space.blocks[id], func(item *Block) bool {
		blocks = append(blocks, item.Clone())
		return len(blocks) < 3
	})

	if len(blocks) == 1 {
		return nil, fmt.Errorf("block not found id: %v", id)
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
		return nil, fmt.Errorf("parent block not found for: %v", id)
	}
	blocks = append(blocks, space.blocks[parent].Clone())

	children, ok := space.children[parent]
	if !ok {
		return nil, fmt.Errorf("block siblings not found id: %v", id)
	}

	children.DescendLessOrEqual(space.blocks[id], func(item *Block) bool {
		blocks = append(blocks, item.Clone())
		return len(blocks) < 3
	})

	if len(blocks) == 1 {
		return nil, fmt.Errorf("block not found id: %v", id)
	}

	return blocks, nil
}

func (ms *MemStore) CreateSpace(space *Space) error {
	if _, ok := ms.spaces[space.ID]; ok {
		return fmt.Errorf("space %v already exists", space.ID)
	}

	ms.spaces[space.ID] = newSpaceStore()
	spaceBlock := NewBlock(space.ID, RootBlockID, "space")
	spaceBlock.Props = NewJsonDoc([]byte(`{"name": "` + space.Name + `"}`))

	err := ms.CreateBlock(&space.ID, spaceBlock)
	if err != nil {
		return err
	}

	return nil
}

// Apply applies transactional changes to the store.
func (ms *MemStore) Apply(tx *Transaction, change *storeChange) error {
	if change == nil {
		return errors.New("cannot apply nil change to store")
	}
	spaceID := &tx.SpaceID

	logrus.Debugf("applying change to store")
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
				return fmt.Errorf("move block not found, %v", block.ID)
			}
			space.RemoveBlock(block.ID)

			storeBlock.ParentID = block.ParentID
			storeBlock.Index = block.Index
			storeBlock.Deleted = block.Deleted
			storeBlock.Erased = block.Erased
			space.AddBlock(storeBlock)
		}

		for _, block := range blockChange.propSet.ToSlice() {
			storeBlock, ok := space.blocks[block.ID]
			if !ok {
				return fmt.Errorf("prop update block not found, %v", block.ID)
			}
			//logrus.Infof("updating props for block %v", block.Props.String())
			storeBlock.Props = block.Props
		}

		// patched blocks should already exist in the store
		for _, block := range blockChange.patched.ToSlice() {
			storeBlock, ok := space.blocks[block.ID]
			if !ok {
				return fmt.Errorf("patch block not found, %v", block.ID)
			}
			//logrus.Infof("patching block %v", block.ID)
			storeBlock.Json = block.Json
		}

		//update the backlinks
		for _, change := range blockChange.linkOps {
			switch change.op {
			case OpTypeUnlink:
				backLinks, ok := space.backLinks[change.childID]
				if !ok {
					backLinks = *NewSet[BlockID]()
				}

				backLinks.Remove(change.parentID)
				space.backLinks[change.childID] = backLinks
			case OpTypeLink:
				backLinks, ok := space.backLinks[change.childID]
				if !ok {
					backLinks = *NewSet[BlockID]()
				}

				backLinks.Add(change.parentID)
				space.backLinks[change.childID] = backLinks
			}
		}

		err := ms.PutTransaction(spaceID, &Transaction{
			ID:      tx.ID,
			SpaceID: tx.SpaceID,
			UserID:  tx.UserID,
			Time:    tx.Time,
			Ops:     tx.Ops,
			changes: change.intoSyncBlocks(),
		})

		if err != nil {
			return err
		}
	}

	//if change.tx != nil {
	//	//TODO implement me
	//	panic("implement me")
	//}

	if change.jsonDocChange != nil {
		//TODO implement me
		panic("implement me")
	}

	return nil
}

func (ms *MemStore) CreateBlock(spaceID *SpaceID, block *Block) error {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		ms.spaces[*spaceID] = space
	}

	space.AddBlock(block)
	ms.blockSpace[block.ID] = *spaceID

	return nil
}

func (ms *MemStore) GetBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	if block, ok := space.blocks[id]; !ok {
		return nil, fmt.Errorf("block %v not found", id)
	} else {
		return block.Clone(), nil
	}
}

func (ms *MemStore) GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	blocks := make([]*Block, 0, len(ids))
	for _, id := range ids {
		if block, ok := space.blocks[id]; ok {
			blocks = append(blocks, block.Clone())
		}
	}
	return blocks, nil
}

func (ms *MemStore) GetAncestorEdges(spaceID *SpaceID, ids []BlockID) ([]blockEdge, error) {
	edges := make([]blockEdge, 0)
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space %v not found", *spaceID)
	}

	for _, id := range ids {
		curr := id
		for {
			parent, ok := space.parents[curr]

			if !ok {
				return nil, fmt.Errorf("non space block %v has no parent", curr)
			}

			if parent == RootBlockID {
				break
			}

			edges = append(edges, blockEdge{parentID: parent, childID: curr})
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
		return nil, fmt.Errorf("space not found: %v", *spaceID)
	}
	return space, nil
}

func (ms *MemStore) GetTransaction(spaceID *SpaceID, id TransactionID) (*Transaction, error) {
	for _, tx := range ms.spaces[*spaceID].txs {
		if tx.ID == id {
			return tx, nil
		}
	}

	return nil, fmt.Errorf("transaction not found: %v", id)
}

func (ms *MemStore) GetNextTransactions(spaceID *SpaceID, id TransactionID, start, limit int) ([]*Transaction, error) {
	space, ok := ms.spaces[*spaceID]
	if !ok {
		return nil, fmt.Errorf("space not found: %v", *spaceID)
	}

	for i, tx := range space.txs {
		if tx.ID == id {
			start, end := i+start+1, i+start+1+limit
			end = min(len(space.txs), end)
			//logrus.Infof("start: %v, end: %v, len: %v", start, end, len(space.txs))
			return space.txs[start:end], nil
		}
	}

	return []*Transaction{}, nil
}

func (ms *MemStore) PutTransaction(spaceID *SpaceID, tx *Transaction) error {
	//logrus.Infof("putting transaction %v", tx.ID)
	space, ok := ms.spaces[*spaceID]
	if !ok {
		space = newSpaceStore()
		ms.spaces[*spaceID] = space
	}

	space.txs = append(space.txs, tx)
	sort.Slice(space.txs, func(i, j int) bool {
		return space.txs[i].Time.After(space.txs[j].Time)
	})
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
