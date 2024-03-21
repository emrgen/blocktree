package blocktree

import (
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/btree"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

// blockTree is a staging ground for loading block from db
type blockTree struct {
	Root     *Block
	Children map[ParentID]*btree.BTreeG[*Block]
}

// newBlockTree creates a block tree from root Block
func newBlockTree(root *Block) *blockTree {
	tree := make(map[ParentID]*btree.BTreeG[*Block])
	tree[root.ID] = btree.NewG(10, blockLessFunc)
	return &blockTree{
		Root:     root,
		Children: tree,
	}
}

// AddEdge add a parent child connection
func (bt *blockTree) AddEdge(parent, child *Block) {
	if tree, ok := bt.Children[parent.ID]; ok {
		tree.ReplaceOrInsert(child)
	}
}

func (bt *blockTree) View() *BlockView {
	return bt.view(bt.Root)
}

func (bt *blockTree) view(parent *Block) *BlockView {
	view := &BlockView{
		Type:     parent.Type,
		ID:       parent.ID,
		ParentID: parent.ParentID,
		Props:    parent.Props,
	}

	if tree, ok := bt.Children[parent.ID]; ok {
		view.Children = make([]*BlockView, 0, tree.Len())
		tree.Ascend(func(b *Block) bool {
			view.Children = append(view.Children, bt.view(b))
			return true
		})
	}

	return view
}

// stageTable is a staging ground for block transaction
type stageTable struct {
	children map[ParentID]*btree.BTreeG[*Block]
	blocks   map[BlockID]*Block
	change   blockChange
	parking  map[BlockID]*Block
}

// NewStageTable creates a new stageTable
func NewStageTable() *stageTable {
	return &stageTable{
		children: make(map[ParentID]*btree.BTreeG[*Block]),
		blocks:   make(map[BlockID]*Block),
		change:   newBlockChange(),
		parking:  make(map[BlockID]*Block),
	}
}

func (st *stageTable) Apply(tx *Transaction) (*blockChange, error) {
	for _, op := range tx.Ops {
		logrus.Debugf("applying op: %s", op.String())
		switch op.Type {
		case OpTypeInsert:
			block, ok := st.parking[op.BlockID]
			if !ok {
				return nil, errors.New("insert block not found")
			}
			// NOTE: space insertion is a special case
			// not need to update index
			if block.Type == "space" {
				logrus.Warnf("space insertion is a special case not done throught transaction")
				continue
			}

			parentId := block.ParentID
			if parentId == uuid.Nil {
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
				if block.Linked {
					st.placeInside(block, op.At.BlockID, Inserted)
				} else {
					return nil, errors.New("invalid position inside for insert block")
				}
			}

			st.unpark(block.ID)
			st.add(block)
			st.change.addInserted(block)
			st.change.addPropSet(parent)
		case OpTypeMove:
			block, ok := st.block(op.BlockID)
			if !ok {
				return nil, fmt.Errorf("move block not found: %v", op.BlockID)
			}
			parentId := block.ParentID
			if parentId == uuid.Nil {
				return nil, fmt.Errorf("old parent id is nil for move block: %v", op.BlockID)
			}
			parent, ok := st.block(parentId)
			//logrus.Infof("existing blocks: %v", st.existingIDs())
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

			st.add(block)
			st.change.addUpdated(block)
			st.change.addPropSet(parent)

		case OpTypeUpdate:
			block, ok := st.block(op.BlockID)
			if !ok {
				return nil, errors.New("update block not found")
			}
			err := block.mergeProps(op.Props)
			if err != nil {
				return nil, err
			}
			logrus.Infof("updated block: %v", block.Props)
			st.change.addPropSet(block)
		case OpTypePatch:
			block, ok := st.block(op.BlockID)
			if !ok {
				return nil, errors.New("patch block not found")
			}
			if block.Json == nil {
				block.Json = DefaultJsonDoc()
			}
			err := block.Json.Apply(op.Patch)
			if err != nil {
				return nil, err
			}
			st.change.addUpdated(block)
		case OpTypeDelete:
			block, ok := st.block(op.BlockID)
			if !ok {
				return nil, errors.New("delete block not found")
			}
			block.Deleted = true
			st.change.addUpdated(block)
		case OpTypeUndelete:
			block, ok := st.block(op.BlockID)
			if !ok {
				return nil, errors.New("undelete block not found")
			}
			block.Deleted = false
			st.change.addUpdated(block)
		case OpTypeErase:
			block, ok := st.block(op.BlockID)
			if !ok {
				return nil, errors.New("erase block not found")
			}
			block.Erased = true
			st.change.addUpdated(block)
		case OpTypeRestore:
			block, ok := st.block(op.BlockID)
			if !ok {
				return nil, errors.New("restore block not found")
			}
			block.Erased = false
			st.change.addUpdated(block)
		}
	}

	return &st.change, nil
}

func (st *stageTable) existingIDs() []BlockID {
	ids := make([]BlockID, 0, len(st.blocks))
	for id := range st.blocks {
		ids = append(ids, id)
	}
	return ids
}

func (st *stageTable) paceAtStart(block *Block, parentID BlockID, action blockChangeType) {
	firstChild, ok := st.firstChild(parentID)
	if ok {
		block.Index = NewBefore(firstChild.Index)
	} else {
		block.Index = DefaultFracIndex()
	}
	block.ParentID = parentID
	st.updateChange(block, action)
}

func (st *stageTable) paceAtEnd(block *Block, parentID BlockID, action blockChangeType) {
	lastChild, ok := st.lastChild(parentID)
	if ok {
		block.Index = NewAfter(lastChild.Index)
	} else {
		block.Index = DefaultFracIndex()
	}
	block.ParentID = parentID
	st.updateChange(block, action)
}

func (st *stageTable) placeBefore(block *Block, nextID BlockID, action blockChangeType) error {
	sibling, err := st.withPrevSibling(nextID)
	if err != nil {
		return err
	}
	if len(sibling) == 0 {
		return errors.New("reference block is not found for place before")
	}

	if len(sibling) == 1 {
		block.Index = NewBefore(sibling[0].Index)
		block.ParentID = sibling[0].ParentID
		st.updateChange(block, action)
	} else if len(sibling) == 2 {
		block.Index, err = NewBetween(sibling[1].Index, sibling[0].Index)
		if err != nil {
			return err
		}
		block.ParentID = sibling[0].ParentID
		st.updateChange(block, action)
	} else {
		return errors.New("invalid sibling count for place before")
	}

	return nil
}

func (st *stageTable) placeAfter(block *Block, prevID BlockID, action blockChangeType) error {
	sibling, err := st.withNextSibling(prevID)
	if err != nil {
		return err
	}
	if len(sibling) == 0 {
		return errors.New("reference block is not found for place after")
	}

	if len(sibling) == 1 {
		block.Index = NewAfter(sibling[0].Index)
		block.ParentID = sibling[0].ParentID
		st.updateChange(block, action)
	} else if len(sibling) == 2 {
		block.Index, err = NewBetween(sibling[0].Index, sibling[1].Index)
		if err != nil {
			return err
		}
		block.ParentID = sibling[0].ParentID
		st.updateChange(block, action)
	} else {
		return errors.New("invalid sibling count for place after")
	}

	return nil
}

func (st *stageTable) placeInside(block *Block, parentID BlockID, action blockChangeType) {
	st.updateChange(block, action)
}

func (st *stageTable) updateChange(block *Block, changeType blockChangeType) {
	switch changeType {
	case Inserted:
		st.change.inserted.Add(block)
	case Updated:
		st.change.updated.Add(block)
	case PropSet:
		st.change.propSet.Add(block)
	}
}

func (st *stageTable) firstChild(parent BlockID) (*Block, bool) {
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

func (st *stageTable) lastChild(parent BlockID) (*Block, bool) {
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
func (st *stageTable) withNextSibling(id BlockID) ([]*Block, error) {
	blocks := make([]*Block, 0, 2)
	block, ok := st.block(id)
	if !ok {
		return blocks, errors.New("block not found")
	}
	if tree, ok := st.children[block.ParentID]; ok {
		if tree.Len() > 0 {
			tree.AscendGreaterOrEqual(block, func(b *Block) bool {
				blocks = append(blocks, b)
				return len(blocks) != 2
			})
		}
	}

	return blocks, nil
}

// withPrevSibling returns [target, Optional[prev]] for a block
func (st *stageTable) withPrevSibling(id BlockID) ([]*Block, error) {
	blocks := make([]*Block, 0, 2)
	block, ok := st.block(id)
	if !ok {
		return blocks, errors.New("block not found")
	}
	if tree, ok := st.children[block.ParentID]; ok {
		if tree.Len() > 0 {
			tree.DescendLessOrEqual(block, func(b *Block) bool {
				blocks = append(blocks, b)
				return len(blocks) != 2
			})
		}
	}

	return blocks, nil
}

// Add adds a block to the table
func (st *stageTable) add(block *Block) {
	st.blocks[block.ID] = block
	if tree, ok := st.children[block.ParentID]; ok {
		tree.ReplaceOrInsert(block)
	} else {
		st.children[block.ParentID] = btree.NewG(10, blockLessFunc)
		st.children[block.ParentID].ReplaceOrInsert(block)
	}
}

// Remove removes a block from the table
func (st *stageTable) remove(block *Block) {
	delete(st.blocks, block.ID)
	if tree, ok := st.children[block.ParentID]; ok {
		tree.Delete(block)
	}
}

func (st *stageTable) block(id BlockID) (*Block, bool) {
	if _, ok := st.blocks[id]; !ok {
		return nil, false
	}
	return st.blocks[id], true
}

// Park a block in the table
func (st *stageTable) park(block *Block) {
	st.parking[block.ID] = block
}

// Unpark a block in the table
func (st *stageTable) parked(id BlockID) (*Block, bool) {
	block, ok := st.parking[id]
	if !ok {
		return nil, false
	}

	return block, true
}

func (st *stageTable) unpark(id BlockID) {
	delete(st.parking, id)
}

func (st *stageTable) contains(id BlockID) bool {
	_, ok := st.blocks[id]
	if ok {
		return true
	}
	_, ok = st.parking[id]
	return ok
}

// blockChange tracks block changes in a transaction
type blockChange struct {
	inserted *Set[*Block]
	updated  *Set[*Block]
	propSet  *Set[*Block]
}

// NewBlockChange creates a new blockChange
func newBlockChange() blockChange {
	return blockChange{
		inserted: NewSet[*Block](),
		updated:  NewSet[*Block](),
		propSet:  NewSet[*Block](),
	}
}

func (bc *blockChange) Inserted() []*Block {
	blocks := make([]*Block, 0, bc.inserted.Cardinality())
	bc.inserted.ForEach(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks
}

func (bc *blockChange) Updated() []*Block {
	blocks := make([]*Block, 0, bc.updated.Cardinality())
	bc.updated.Difference(bc.inserted).ForEach(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks
}

func (bc *blockChange) PropSet() []*Block {
	blocks := make([]*Block, 0, bc.propSet.Cardinality())
	bc.propSet.ForEach(func(item *Block) bool {
		blocks = append(blocks, item)
		return true
	})

	return blocks
}

func (bc *blockChange) addInserted(id *Block) {
	bc.inserted.Add(id)
}

func (bc *blockChange) addUpdated(id *Block) {
	bc.updated.Add(id)
}

func (bc *blockChange) addPropSet(id *Block) {
	bc.propSet.Add(id)
}

func (bc *blockChange) empty() bool {
	return bc.inserted.Cardinality() == 0 && bc.updated.Cardinality() == 0 && bc.propSet.Cardinality() == 0
}

type blockChangeType string

const (
	Inserted blockChangeType = "inserted"
	Updated  blockChangeType = "updated" // includes move, link, unlink, delete, erase
	PropSet  blockChangeType = "prop_set"
)

// moveTree helps to track block cycle in a space
type moveTree struct {
	spaceId   SpaceID
	blocks    Set[BlockID]
	backEdges map[BlockID]BlockID
}

// newMoveTree creates a new moveTree
func newMoveTree(spaceId SpaceID) *moveTree {
	blocks := NewSet(spaceId)
	return &moveTree{
		spaceId:   spaceId,
		blocks:    *blocks,
		backEdges: make(map[BlockID]BlockID),
	}
}

// Move moves a block to a new parent
func (mt *moveTree) Move(child, parent BlockID) error {
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
func (mt *moveTree) addEdge(child, parent BlockID) {
	mt.blocks.Add(parent)
	mt.blocks.Add(child)
	mt.backEdges[child] = parent
}

// getParent returns the parent of a block
func (mt *moveTree) getParent(block BlockID) (*BlockID, bool) {
	if parent, ok := mt.backEdges[block]; ok {
		return &parent, true
	}
	return nil, false
}

func (mt *moveTree) contains(block BlockID) bool {
	return mt.blocks.Contains(block)
}

func (mt *moveTree) print() {
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

type blockEdge struct {
	parentID BlockID
	childID  BlockID
}
