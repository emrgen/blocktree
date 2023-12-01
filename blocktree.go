package blocktree

import (
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

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
		bt.change.inserted.Add(block)
	case Updated:
		bt.change.updated.Add(block)
	case PropSet:
		bt.change.propSet.Add(block)
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
	if tree, ok := bt.children[*block.ParentID]; ok {
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
	if tree, ok := bt.children[*block.ParentID]; ok {
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
	if block.ParentID != nil {
		if tree, ok := bt.children[*block.ParentID]; ok {
			tree.ReplaceOrInsert(block)
		} else {
			bt.children[*block.ParentID] = btree.NewG(10, blockLessFunc)
			bt.children[*block.ParentID].ReplaceOrInsert(block)
		}
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
	inserted *Set[*Block]
	updated  *Set[*Block]
	propSet  *Set[*Block]
}

// NewBlockChange creates a new BlockChange
func newBlockChange() BlockChange {
	return BlockChange{
		inserted: NewSet[*Block](),
		updated:  NewSet[*Block](),
		propSet:  NewSet[*Block](),
	}
}

func (bc *BlockChange) Inserted() []*Block {
	blocks := make([]*Block, 0, bc.inserted.Cardinality())
	bc.inserted.ForEach(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks
}

func (bc *BlockChange) Updated() []*Block {
	blocks := make([]*Block, 0, bc.updated.Cardinality())
	bc.updated.Difference(bc.inserted).ForEach(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks
}

func (bc *BlockChange) PropSet() []*Block {
	blocks := make([]*Block, 0, bc.propSet.Cardinality())
	bc.propSet.ForEach(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks
}

func (bc *BlockChange) addInserted(id *Block) {
	bc.inserted.Add(id)
}

func (bc *BlockChange) addUpdated(id *Block) {
	bc.updated.Add(id)
}

func (bc *BlockChange) addPropSet(id *Block) {
	bc.propSet.Add(id)
}

func (bc *BlockChange) empty() bool {
	return bc.inserted.Cardinality() == 0 && bc.updated.Cardinality() == 0 && bc.propSet.Cardinality() == 0
}

type BlockChangeType string

const (
	Inserted BlockChangeType = "inserted"
	Updated  BlockChangeType = "updated" // includes move, link, unlink, delete, erase
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
func (mt *MoveTree) Move(child, parent BlockID) error {
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
		if parentId, ok = mt.backEdges[parentId]; ok {
			if parentId == mt.spaceId {
				break
			}

			if parentId == child {
				return ErrCreatesCycle
			}
			if visited.Contains(parentId) {
				return ErrDetectedCycle
			}
			visited.Add(parentId)
		} else {
			break
		}
	}

	// move child to new parent
	delete(mt.backEdges, child)
	mt.addEdge(child, parent)

	return nil
}

// addEdge adds a parent child connection
// does not check for cycles
func (mt *MoveTree) addEdge(child, parent BlockID) {
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

func (mt *MoveTree) print() {
	children := make(map[BlockID][]BlockID)
	for child, parent := range mt.backEdges {
		if _, ok := children[parent]; !ok {
			children[parent] = make([]BlockID, 0)
		}
		children[parent] = append(children[parent], child)
	}

	tree := treeprint.New()
	space := tree.AddBranch(mt.spaceId.String())
	traverse(mt.spaceId, children, space)

	logrus.Infof("%v", tree.String())
}

func traverse(parent BlockID, children map[BlockID][]BlockID, space treeprint.Tree) {
	for _, child := range children[parent] {
		branch := space.AddBranch(child.String())
		traverse(child, children, branch)
	}
}

type BlockEdge struct {
	parentID BlockID
	childID  BlockID
}
