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

func (st *StageTable) Apply(tx *Transaction) (*BlockChange, error) {
	for _, op := range tx.Ops {
		logrus.Infof("applying op: %s", op.String())
		switch op.Type {
		case OpTypeInsert:
			block, ok := st.parking[op.BlockID]
			if !ok {
				return nil, errors.New("insert block not found")
			}
			// NOTE: space insertion is a special case
			// not need to update index
			if block.Type == "space" {
				st.change.addInserted(block)
				continue
			}

			parentId := block.ParentID
			if parentId == nil {
				return nil, errors.New("parent id is nil for insert block")
			}
			parent, ok := st.block(op.At.BlockID)
			if !ok {
				return nil, errors.New("parent block not found for insert at start")
			}
			switch op.At.Position {
			case PositionStart:
				st.paceAtStart(block, op.At.BlockID, Inserted)
			case PositionEnd:
				st.paceAtEnd(block, op.At.BlockID, Inserted)
			case PositionBefore:
				err := st.placeBefore(block, op.At.BlockID, Inserted)
				if err != nil {
					return nil, err
				}
			case PositionAfter:
				err := st.placeAfter(block, op.At.BlockID, Inserted)
				if err != nil {
					return nil, err
				}
			case PositionInside:
				return nil, errors.New("invalid position inside for insert block")
			}

			st.change.addInserted(block)
			st.change.addUpdated(parent)
		case OpTypeMove:
			block, ok := st.parking[op.BlockID]
			if !ok {
				return nil, errors.New("move block not found")
			}
			parentId := block.ParentID
			if parentId == nil {
				return nil, errors.New("old parent id is nil for move block")
			}
			parent, ok := st.block(*parentId)
			if !ok {
				return nil, errors.New("old parent block not found for move block")
			}

			// remove block from its current position
			// to ensure that the block (subtree) is not in the tree
			// the subtree nodes are still in the table but the connection is removed
			st.remove(block)

			switch op.At.Position {
			case PositionStart:
				st.paceAtStart(block, op.At.BlockID, Updated)
			case PositionEnd:
				st.paceAtEnd(block, op.At.BlockID, Updated)
			case PositionBefore:
				err := st.placeBefore(block, op.At.BlockID, Updated)
				if err != nil {
					return nil, err
				}
			case PositionAfter:
				err := st.placeAfter(block, op.At.BlockID, Updated)
				if err != nil {
					return nil, err
				}
			case PositionInside:
				return nil, errors.New("invalid position inside for move block")
			}

			st.change.addUpdated(block)
			st.change.addUpdated(parent)

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
		}
	}

	return &st.change, nil
}

func (st *StageTable) paceAtStart(block *Block, parentID BlockID, action BlockChangeType) {
	firstChild, ok := st.firstChild(parentID)
	if ok {
		block.Index = NewBefore(firstChild.Index)
	} else {
		block.Index = DefaultFracIndex()
	}
	st.updateChange(block, action)
}

func (st *StageTable) paceAtEnd(block *Block, parentID BlockID, action BlockChangeType) {
	lastChild, ok := st.lastChild(parentID)
	if ok {
		block.Index = NewAfter(lastChild.Index)
	} else {
		block.Index = DefaultFracIndex()
	}
	st.updateChange(block, action)
}

func (st *StageTable) placeBefore(block *Block, nextID BlockID, action BlockChangeType) error {
	sibling, err := st.withPrevSibling(nextID)
	if err != nil {
		return err
	}
	if len(sibling) == 0 {
		return errors.New("reference block is not found for place before")
	}

	if len(sibling) == 1 {
		block.Index = NewBefore(sibling[0].Index)
		st.updateChange(block, action)
	} else if len(sibling) == 2 {
		block.Index, err = NewBetween(sibling[1].Index, sibling[0].Index)
		if err != nil {
			return err
		}
		st.updateChange(block, action)
	} else {
		return errors.New("invalid sibling count for place before")
	}

	return nil
}

func (st *StageTable) placeAfter(block *Block, prevID BlockID, opType BlockChangeType) error {
	sibling, err := st.withNextSibling(prevID)
	if err != nil {
		return err
	}
	if len(sibling) == 0 {
		return errors.New("reference block is not found for place after")
	}

	if len(sibling) == 1 {
		block.Index = NewAfter(sibling[0].Index)
		st.updateChange(block, opType)
	} else if len(sibling) == 2 {
		block.Index, err = NewBetween(sibling[0].Index, sibling[1].Index)
		if err != nil {
			return err
		}
		st.updateChange(block, opType)
	} else {
		return errors.New("invalid sibling count for place after")
	}

	return nil
}

func (st *StageTable) updateChange(block *Block, changeType BlockChangeType) {
	switch changeType {
	case Inserted:
		st.change.inserted.Add(block)
	case Updated:
		st.change.updated.Add(block)
	case PropSet:
		st.change.propSet.Add(block)
	}
}

func (st *StageTable) firstChild(parent BlockID) (*Block, bool) {
	if tree, ok := st.children[parent]; ok {
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

func (st *StageTable) lastChild(parent BlockID) (*Block, bool) {
	if tree, ok := st.children[parent]; ok {
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
func (st *StageTable) withNextSibling(id BlockID) ([]*Block, error) {
	blocks := make([]*Block, 0, 2)
	block, ok := st.block(id)
	if !ok {
		return blocks, errors.New("block not found")
	}
	if tree, ok := st.children[*block.ParentID]; ok {
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
func (st *StageTable) withPrevSibling(id BlockID) ([]*Block, error) {
	blocks := make([]*Block, 0, 2)
	block, ok := st.block(id)
	if !ok {
		return blocks, errors.New("block not found")
	}
	if tree, ok := st.children[*block.ParentID]; ok {
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
func (st *StageTable) add(block *Block) {
	st.blocks[block.ID] = block
	if block.ParentID != nil {
		if tree, ok := st.children[*block.ParentID]; ok {
			tree.ReplaceOrInsert(block)
		} else {
			st.children[*block.ParentID] = btree.NewG(10, blockLessFunc)
			st.children[*block.ParentID].ReplaceOrInsert(block)
		}
	}
}

// Remove removes a block from the table
func (st *StageTable) remove(block *Block) {
	delete(st.blocks, block.ID)
	if block.ParentID != nil {
		if tree, ok := st.children[*block.ParentID]; ok {
			tree.Delete(block)
		}
	}
}

func (st *StageTable) block(id BlockID) (*Block, bool) {
	if _, ok := st.blocks[id]; !ok {
		return nil, false
	}
	return st.blocks[id], true
}

// Park a block in the table
func (st *StageTable) park(block *Block) {
	st.parking[block.ID] = block
}

// Unpark a block in the table
func (st *StageTable) parked(id BlockID) (*Block, bool) {
	block, ok := st.parking[id]
	if !ok {
		return nil, false
	}

	return block, true
}

func (st *StageTable) unpark(id BlockID) {
	delete(st.parking, id)
}

func (st *StageTable) contains(id BlockID) bool {
	_, ok := st.blocks[id]
	if ok {
		return true
	}
	_, ok = st.parking[id]
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
