package blocktree

import (
	"github.com/google/uuid"
)

type BlockID = uuid.UUID
type ParentID = uuid.UUID
type ChildID = uuid.UUID
type SpaceID = uuid.UUID

type Space struct {
	ID   SpaceID
	Name string
}

type BlockProps = map[string]interface{}

type BlockView struct {
	Type     string
	ID       uuid.UUID
	ParentID *uuid.UUID
	Props    BlockProps
	Children []*BlockView
	Deleted  bool
	Erased   bool
}

type Block struct {
	Type     string
	ID       BlockID
	ParentID *ParentID
	Index    *FracIndex
	Props    BlockProps
	Deleted  bool
	Erased   bool
}

func NewBlock(blockID BlockID, parentID *ParentID, blockType string) *Block {
	return &Block{
		Type:     blockType,
		ID:       blockID,
		ParentID: parentID,
		Index:    DefaultFracIndex(),
	}
}

func (b *Block) Clone() *Block {
	return &Block{
		Type:     b.Type,
		ID:       b.ID,
		ParentID: b.ParentID,
		Index:    b.Index.Clone(),
		Props:    b.Props,
		Deleted:  b.Deleted,
		Erased:   b.Erased,
	}
}

// Less allows btree entry
func (b *Block) Less(other *Block) bool {
	return b.Index.Compare(other.Index) < 0
}

func blockLessFunc(a, b *Block) bool {
	return a.Less(b)
}
