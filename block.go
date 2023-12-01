package blocktree

import (
	"errors"
	"github.com/deckarep/golang-set/v2"
	"github.com/google/btree"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type BlockID = uuid.UUID
type ParentID = uuid.UUID
type ChildID = uuid.UUID

type BlockProps = map[string]interface{}

type BlockView struct {
	Type     string
	ID       uuid.UUID
	ParentID uuid.UUID
	Props    BlockProps
	Children []*BlockView
	Deleted  bool
	Erased   bool
}

type Block struct {
	Type     string
	ID       uuid.UUID
	ParentID uuid.UUID
	Index    *FracIndex
	Props    BlockProps
	Deleted  bool
	Erased   bool
}

func NewBlock(parentID, blockID ParentID, blockType string) *Block {
	return &Block{
		Type:     blockType,
		ID:       blockID,
		ParentID: parentID,
		Index:    DefaultFracIndex(),
	}
}

// Less allows btree entry
func (b *Block) Less(other *Block) bool {
	return b.Index.Compare(other.Index) < 0 || b.ID.String() < other.ID.String()
}

func blockLessFunc(a, b *Block) bool {
	return a.Less(b)
}

// BlockTree is a staging ground for loading block from db
type BlockTree struct {
	Root     *Block
	Children map[ParentID]*btree.BTreeG[*Block]
}

// NewBlockTree creates a block tree from root Block
func NewBlockTree(root *Block) *BlockTree {
	tree := make(map[ParentID]*btree.BTreeG[*Block])
	tree[root.ID] = btree.NewG(10, blockLessFunc)
	return &BlockTree{
		Root:     root,
		Children: tree,
	}
}

// AddEdge add a parent child connection
func (bt *BlockTree) AddEdge(parent, child *Block) {
	if tree, ok := bt.Children[parent.ID]; ok {
		tree.ReplaceOrInsert(child)
	}
}

func (bt *BlockTree) View() *BlockView {
	return bt.view(bt.Root)
}

func (bt *BlockTree) view(parent *Block) *BlockView {
	blockView := &BlockView{
		Type:     parent.Type,
		ID:       parent.ID,
		ParentID: parent.ParentID,
		Props:    parent.Props,
	}

	if tree, ok := bt.Children[parent.ID]; ok {
		blockView.Children = make([]*BlockView, 0, tree.Len())
		tree.Ascend(func(b *Block) bool {
			blockView.Children = append(blockView.Children, bt.view(b))
			return true
		})
	}

	return blockView
}

// StageTable is a staging ground for block transaction
type StageTable struct {
	children map[ParentID]*btree.BTreeG[*Block]
	blocks   map[BlockID]*Block
	change   BlockChange
	parking  map[BlockID]*Block
}

// NewStageTable creates a new StageTable
func NewStageTable() *StageTable {
	return &StageTable{
		children: make(map[ParentID]*btree.BTreeG[*Block]),
		blocks:   make(map[BlockID]*Block),
		change:   newBlockChange(),
		parking:  make(map[BlockID]*Block),
	}
}

func (bt *StageTable) Apply(tx *Transaction) (*BlockChange, error) {
	// load the referenced blocks
	//blocks := NewSet[BlockID]()
	stage := NewStageTable()
	for _, op := range tx.Ops {
		logrus.Infof("applying op: %v", op)
		//switch op.Type {
		//case OpTypeInsert:
		//	block := &Block{
		//		Type:     op.BlockType,
		//		ID:       op.BlockID,
		//		ParentID: op.ParentID,
		//		Index:    op.Index,
		//		Props:    op.Props,
		//		Deleted:  false,
		//		Erased:   false,
		//	}
		//	blocks[block.ID] = block
		//	stage.add(block)
		//case OpTypeUpdate:
		//	if block, ok := blocks[op.BlockID]; ok {
		//		block.Props = op.Props
		//		stage.updateChange(block, Updated)
		//	}
		//case OpTypePatch:
		//	if block, ok := blocks[op.BlockID]; ok {
		//		block.Props = op.Props
		//		stage.updateChange(block, PropSet)
		//	}
		//case OpTypeMove:
		//	if block, ok := blocks[op.BlockID]; ok {
		//		block.ParentID = op.ParentID
		//		stage.updateChange(block, Updated)
		//	}
		//case OpTypeLink:
		//	if block, ok := blocks[op.BlockID]; ok {
		//		block.ParentID = op.ParentID
		//		stage.updateChange(block, Updated)
		//	}
		//case OpTypeUnlink:
		//	if block, ok := blocks[op.BlockID]; ok {
		//		block.ParentID = op.ParentID
		//		stage.updateChange(block, Updated)
		//	}
		//case OpTypeDelete:
		//	if block, ok := blocks[op.BlockID]; ok {
		//		block.Deleted = true
		//		stage.updateChange(block, Updated)
		//	}
		//case OpTypeErase:
		//	if block, ok := blocks[op.BlockID]; ok {
		//		block.Erased = true
		//		stage.updateChange(block, Updated)
		//	}
		//}
	}

	// apply changes to the tree
	//for _, block := range blocks {
	//	if block.Deleted {
	//		if tree, ok := stage.children[block.ParentID]; ok {
	//			tree.Delete(block)
	//		}
	//	} else {
	//		stage.add(block)
	//	}
	//}
	//
	//// update changes
	//for _, block := range blocks {
	//	if block.Deleted {
	//		stage.updateChange(block, Updated)
	//	}
	//}

	// update

	return &stage.change, nil
}

func (bt *StageTable) updateChange(block *Block, changeType BlockChangeType) {
	switch changeType {
	case Inserted:
		bt.change.inserted.Add(block.ID)
	case Updated:
		bt.change.updated.Add(block.ID)
	case PropSet:
		bt.change.propSet.Add(block.ID)
	}
}

func (bt *StageTable) firstChild(parent BlockID) (*Block, bool) {
	if tree, ok := bt.children[parent]; ok {
		if tree.Len() > 0 {
			var first *Block
			tree.Ascend(func(b *Block) bool {
				first = b
				return false
			})
			return first, true
		}
	}
	return nil, false
}

func (bt *StageTable) lastChild(parent BlockID) (*Block, bool) {
	if tree, ok := bt.children[parent]; ok {
		if tree.Len() > 0 {
			var last *Block
			tree.Descend(func(b *Block) bool {
				last = b
				return false
			})
			return last, true
		}
	}
	return nil, false
}

// withNextSibling returns [target, Optional[next]] for a block
func (bt *StageTable) withNextSibling(id BlockID) ([]*Block, error) {
	blocks := make([]*Block, 0, 2)
	block, ok := bt.block(id)
	if !ok {
		return blocks, errors.New("block not found")
	}
	if tree, ok := bt.children[block.ParentID]; ok {
		if tree.Len() > 0 {
			tree.AscendGreaterOrEqual(block, func(b *Block) bool {
				blocks = append(blocks, b)
				if len(blocks) == 2 {
					return false
				}
				return true
			})
		}
	}

	return blocks, nil
}

// withPrevSibling returns [target, Optional[prev]] for a block
func (bt *StageTable) withPrevSibling(id BlockID) ([]*Block, error) {
	blocks := make([]*Block, 0, 2)
	block, ok := bt.block(id)
	if !ok {
		return blocks, errors.New("block not found")
	}
	if tree, ok := bt.children[block.ParentID]; ok {
		if tree.Len() > 0 {
			tree.DescendLessOrEqual(block, func(b *Block) bool {
				blocks = append(blocks, b)
				if len(blocks) == 2 {
					return false
				}
				return true
			})
		}
	}

	return blocks, nil
}

// Add adds a block to the table
func (bt *StageTable) add(block *Block) {
	bt.blocks[block.ID] = block
	if tree, ok := bt.children[block.ParentID]; ok {
		tree.ReplaceOrInsert(block)
	} else {
		bt.children[block.ParentID] = btree.NewG(10, blockLessFunc)
		bt.children[block.ParentID].ReplaceOrInsert(block)
	}
}

func (bt *StageTable) block(id BlockID) (*Block, bool) {
	if _, ok := bt.blocks[id]; !ok {
		return nil, false
	}
	return bt.blocks[id], true
}

// Park a block in the table
func (bt *StageTable) park(block *Block) {
	bt.parking[block.ID] = block
}

// Unpark a block in the table
func (bt *StageTable) parked(id BlockID) (*Block, bool) {
	block, ok := bt.parking[id]
	if !ok {
		return nil, false
	}

	return block, true
}

func (bt *StageTable) unpark(id BlockID) {
	delete(bt.parking, id)
}

func (bt *StageTable) contains(id BlockID) bool {
	_, ok := bt.blocks[id]
	if ok {
		return true
	}
	_, ok = bt.parking[id]
	if ok {
		return true
	}
	return false
}

// BlockChange tracks block changes in a transaction
type BlockChange struct {
	inserted mapset.Set[BlockID]
	updated  mapset.Set[BlockID]
	propSet  mapset.Set[BlockID]
}

// NewBlockChange creates a new BlockChange
func newBlockChange() BlockChange {
	return BlockChange{}
}

func (bc *BlockChange) addInserted(id BlockID) {
	if bc.inserted == nil {
		bc.inserted = mapset.NewSet(id)
	} else {
		bc.inserted.Add(id)
	}
}

func (bc *BlockChange) addUpdated(id BlockID) {
	if bc.updated == nil {
		bc.updated = mapset.NewSet(id)
	} else {
		bc.updated.Add(id)
	}
}

func (bc *BlockChange) addPropSet(id BlockID) {
	if bc.propSet == nil {
		bc.propSet = mapset.NewSet(id)
	} else {
		bc.propSet.Add(id)
	}
}

func (bc *BlockChange) empty() bool {
	count := 0
	if bc.inserted != nil {
		count += bc.inserted.Cardinality()
	}
	if bc.updated != nil {
		count += bc.updated.Cardinality()
	}
	if bc.propSet != nil {
		count += bc.propSet.Cardinality()
	}
	return count == 0
}

type BlockChangeType string

const (
	Inserted BlockChangeType = "inserted"
	Updated  BlockChangeType = "updated"
	PropSet  BlockChangeType = "prop_set"
)

// MoveTree helps to track block cycle in a space
type MoveTree struct {
	spaceId   SpaceID
	blocks    Set[BlockID]
	backEdges map[BlockID]BlockID
}

// NewMoveTree creates a new MoveTree
func NewMoveTree(spaceId SpaceID) *MoveTree {
	blocks := NewSet(spaceId)
	return &MoveTree{
		spaceId:   spaceId,
		blocks:    *blocks,
		backEdges: make(map[BlockID]BlockID),
	}
}

// Move moves a block to a new parent
func (mt *MoveTree) Move(parent, child BlockID) error {
	if mt.spaceId == child {
		return errors.New("cannot move space block")
	}

	if parent == child {
		return errors.New("cannot move block to itself")
	}

	if !mt.blocks.Contains(child) {
		return errors.New("child block not found")
	}

	if !mt.blocks.Contains(parent) {
		return errors.New("parent block not found")
	}

	// if child_id is already a child of parent_id, then the move is not needed
	currParent, ok := mt.backEdges[child]
	if !ok {
		return errors.New("child block not found")
	}
	if currParent == parent {
		return nil
	}

	// if child is an ancestor of parent, then the move would create a cycle
	parentId := parent
	visited := mapset.NewSet(parentId)
	for {
		if blockId, ok := mt.backEdges[parentId]; ok {
			if blockId == child {
				return ErrCreatesCycle
			}
			if visited.Contains(blockId) {
				return ErrDetectedCycle
			}
			visited.Add(blockId)
		} else {
			break
		}
	}

	// move child to new parent
	delete(mt.backEdges, child)
	mt.addEdge(parent, child)

	return nil
}

// addEdge adds a parent child connection
func (mt *MoveTree) addEdge(parent, child BlockID) {
	mt.blocks.Add(parent)
	mt.blocks.Add(child)
	mt.backEdges[child] = parent
}

// getParent returns the parent of a block
func (mt *MoveTree) getParent(block BlockID) (*BlockID, bool) {
	if parent, ok := mt.backEdges[block]; ok {
		return &parent, true
	}
	return nil, false
}

func (mt *MoveTree) contains(block BlockID) bool {
	return mt.blocks.Contains(block)
}

type BlockEdge struct {
	parentID BlockID
	childID  BlockID
}
