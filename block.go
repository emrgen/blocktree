package blocktree

import (
	"errors"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/google/btree"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

var (
	RootBlockID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
)

type BlockID = uuid.UUID
type ParentID = uuid.UUID
type ChildID = uuid.UUID
type SpaceID = uuid.UUID

type Space struct {
	ID   SpaceID
	Name string
}

func NewSpace(spaceID SpaceID, name string) *Space {
	return &Space{
		ID:   spaceID,
		Name: name,
	}
}

type BlockProps = map[string]interface{}

type BlockView struct {
	Type     string
	ID       uuid.UUID
	ParentID uuid.UUID
	Props    BlockProps
	Children []*BlockView
	Linked   []*BlockView
	Deleted  bool
	Erased   bool
}

func BlockViewFromBlock(block *Block) *BlockView {
	return &BlockView{
		Type:     block.Type,
		ID:       block.ID,
		ParentID: block.ParentID,
		Props:    block.Props,
		Deleted:  block.Deleted,
		Erased:   block.Erased,
	}
}

func BlockViewFromBlocks(rootID BlockID, blocks []*Block) (*BlockView, error) {
	var root *BlockView
	children := make(map[BlockID]*btree.BTreeG[*Block])
	linked := make(map[BlockID]*Set[*Block])

	for _, block := range blocks {
		if block.ID == rootID {
			root = BlockViewFromBlock(block)
		}

		if block.Linked {
			if _, ok := linked[block.ParentID]; !ok {
				linked[block.ParentID] = NewSet[*Block]()
			}
			linked[block.ParentID].Add(block)
		} else {
			if _, ok := children[block.ParentID]; !ok {
				children[block.ParentID] = btree.NewG(10, blockLessFunc)
			}
			logrus.Infof("inserting block %v to parent %v", block.ID, block.ParentID)
			children[block.ParentID].ReplaceOrInsert(block)
		}
	}

	if root == nil {
		return nil, errors.New("root block not found")
	}

	// build tree function
	var build func(*BlockView)
	build = func(block *BlockView) {
		if children, ok := children[block.ID]; ok {
			block.Children = make([]*BlockView, 0)
			children.Ascend(func(item *Block) bool {
				child := BlockViewFromBlock(item)
				block.Children = append(block.Children, child)
				build(child)
				return true
			})
		}

		if linked, ok := linked[block.ID]; ok {
			block.Linked = make([]*BlockView, 0)
			for _, item := range linked.ToSlice() {
				child := BlockViewFromBlock(item)
				block.Linked = append(block.Linked, child)
				build(child)
			}
		}
	}

	// build tree
	build(root)

	return root, nil
}

func (b *BlockView) Print() {
	var build func(treeprint.Tree, *BlockView)
	build = func(tree treeprint.Tree, block *BlockView) {
		v := tree.AddNode(block.ID.String())

		if len(block.Children) > 0 {
			children := v.AddBranch("children")
			for _, child := range block.Children {
				build(children, child)
			}
		}

		if len(block.Linked) > 0 {
			linked := v.AddBranch("linked")
			for _, child := range block.Linked {
				build(linked, child)
			}
		}
	}

	tree := treeprint.New()
	build(tree, b)

	logrus.Infof("tree: %v", tree.String())
}

type Block struct {
	Type        string
	Table       string
	ID          BlockID
	ParentID    ParentID
	Index       *FracIndex
	Props       BlockProps
	Json        []byte
	Deleted     bool
	Erased      bool
	Linked      bool // linked blocks
	UpdateFlags uint32
}

func NewBlock(blockID BlockID, parentID ParentID, blockType string) *Block {
	return &Block{
		Type:     blockType,
		ID:       blockID,
		ParentID: parentID,
		Index:    DefaultFracIndex(),
	}
}

func (b *Block) mergeProps(props []OpProp) {
	var merge func(map[string]interface{}, []string, interface{}) interface{}
	merge = func(props map[string]interface{}, keys []string, value interface{}) interface{} {
		if len(keys) == 0 {
			return value
		}

		if props == nil {
			props = make(map[string]interface{})
		}

		if _, ok := props[keys[0]]; !ok {
			props[keys[0]] = make(map[string]interface{})
		}

		merge(props[keys[0]].(map[string]interface{}), keys[1:], value)
		return props
	}

	for _, prop := range props {
		merge(b.Props, prop.Path, prop.Value)
	}
}

func (b *Block) mergeJson(patch []byte) error {
	content, err := jsonpatch.MergePatch(b.Json, patch)
	if err != nil {
		return err
	}
	b.Json = content
	return nil
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
		Linked:   b.Linked,
	}
}

// Less allows btree entry
func (b *Block) Less(other *Block) bool {
	// linked blocks will have clashing index causing btree to block overwrite by id
	if b.Index.Equals(other.Index) {
		return b.ID.String() < other.ID.String()
	}
	return b.Index.Compare(other.Index) < 0
}

func blockLessFunc(a, b *Block) bool {
	return a.Less(b)
}
