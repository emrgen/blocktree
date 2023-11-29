package blocktree

import (
	"errors"
	"github.com/deckarep/golang-set/v2"
	"github.com/google/btree"
	"github.com/google/uuid"
)

type BlockID = uuid.UUID
type ParentID = uuid.UUID
type ChildID = uuid.UUID

type BlockProps struct {
}

type BlockView struct {
	Type     string
	ID       uuid.UUID
	ParentID uuid.UUID
	Props    BlockProps
	Children []*BlockView
}

type Block struct {
	Type     string
	ID       uuid.UUID
	ParentID uuid.UUID
	Index    *FracIndex
	Props    BlockProps
}

// Less allows btree entry
func (b *Block) Less(other *Block) bool {
	return b.Index.Compare(other.Index) < 0
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

// BlockTable is a staging ground for block transaction
type BlockTable struct {
	children map[ParentID]*btree.BTreeG[*Block]
	change   BlockChange
	parking  map[BlockID]*Block
}

type BlockChange struct {
	Inserted []*Block
	Updated  []*Block
	Modified []*Block
}

// MoveTree helps to track block cycle in a space
type MoveTree struct {
	spaceId   SpaceID
	blocks    mapset.Set[BlockID]
	backEdges map[BlockID]BlockID
}

func NewMoveTree(spaceId SpaceID) *MoveTree {
	return &MoveTree{
		spaceId:   spaceId,
		blocks:    mapset.NewSet(spaceId),
		backEdges: make(map[BlockID]BlockID),
	}
}

func (mt *MoveTree) Move(parent, child BlockID) error {
	if mt.spaceId == child {
		return errors.New("cannot move space block")
	}

	if parent == child {
		return errors.New("cannot move block to itself")
	}

	if !mt.contains(child) {
		return errors.New("child block not found")
	}

	if !mt.contains(parent) {
		return errors.New("parent block not found")
	}

	// if child_id is already a child of parent_id, then the move is not needed
	if mt.backEdges[child] == parent {
		return nil
	}

	// if child is an ancestor of parent, then the move would create a cycle
	parentId := parent
	for {
		if blockId, ok := mt.backEdges[parentId]; ok {
			if blockId == child {
				return errors.New("move would create a cycle")
			}
		} else {
			break
		}
	}

	oldParent := mt.backEdges[child]
	mt.removeEdge(oldParent, child)
	mt.addEdge(parent, child)

	return nil
}

func (mt *MoveTree) addEdge(parent, child BlockID) {
	mt.blocks.Add(parent)
	mt.blocks.Add(child)
	mt.backEdges[child] = parent
}

func (mt *MoveTree) removeEdge(parent, child BlockID) {
	delete(mt.backEdges, child)
}

func (mt *MoveTree) contains(block BlockID) bool {
	return mt.blocks.Contains(block)
}
