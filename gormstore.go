package blocktree

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	_ Store = (*GormStore)(nil)
)

// GormStore is a blocktree store backed by a gorm database.
type GormStore struct {
	db *gorm.DB
}

// NewGormStore creates a new GormStore. It requires a gorm database.
func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (g GormStore) GetLatestTransaction(spaceID *SpaceID) (*Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) Apply(tx *Transaction, change *storeChange) error {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) CreateSpace(space *Space) error {
	model := space.toGormSpace()
	return g.db.Create(model).Error
}

func (g GormStore) GetBlockSpaceID(id *BlockID) (*SpaceID, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) CreateBlock(spaceID *SpaceID, block *Block) error {
	model := block.toGormBlock()
	return g.db.Create(model).Error
}

func (g GormStore) GetBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	var model gormBlock
	res := g.db.First(&model, "id = ?", id)
	if res.Error != nil {
		return nil, res.Error
	}

	return model.toBlock()
}

func (g GormStore) GetChildrenBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	var blocks []*gormBlock
	res := g.db.Where("parent_id = ?", id).Find(&blocks)
	if res.Error != nil {
		return nil, res.Error
	}

	ret := make([]*Block, len(blocks))
	for i, b := range blocks {
		e, err := b.toBlock()
		if err != nil {
			return nil, err
		}
		ret[i] = e
	}

	return ret, nil
}

func (g GormStore) GetLinkedBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetDescendantBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetParentBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error) {
	var blocks []*gormBlock
	res := g.db.Where("id IN (?)", ids).Find(&blocks)
	if res.Error != nil {
		return nil, res.Error
	}

	ret := make([]*Block, len(blocks))
	for i, b := range blocks {
		block, err := b.toBlock()
		if err != nil {
			return nil, err
		}
		ret[i] = block
	}

	return ret, nil
}

func (g GormStore) GetWithFirstChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetWithLastChildBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetParentWithNextBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetParentWithPrevBlock(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetAncestorEdges(spaceID *SpaceID, id []BlockID) ([]blockEdge, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetTransaction(spaceID *SpaceID, id TransactionID) (*Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) PutTransaction(spaceID *SpaceID, tx *Transaction) error {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) ApplyChange(space *SpaceID, change *storeChange) error {
	//TODO implement me
	panic("implement me")
}

// gormSpace is a space in gorm database.
type gormSpace struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key"`
	Name string    `gorm:"not null"`
}

//func (s *gormSpace) toSpace() *Space {
//	return &Space{
//		ID:   s.ID,
//		Name: s.Name,
//	}
//}

func (s *Space) toGormSpace() *gormSpace {
	return &gormSpace{
		ID:   s.ID,
		Name: s.Name,
	}
}

type gormBlock struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key"`
	Type     string    `gorm:"not null"`
	ParentID uuid.UUID `gorm:"type:uuid;not null"`
	Index    string    `gorm:"not null"`
	Deleted  bool      `gorm:"not null"`
	Erased   bool      `gorm:"not null"`
}

func (b *gormBlock) toBlock() (*Block, error) {
	block := Block{
		ID:       b.ID,
		ParentID: b.ParentID,
		Type:     b.Type,
		Deleted:  b.Deleted,
		Erased:   b.Erased,
	}

	if b.Index != "" {
		id := FracIndexFromBytes([]byte(b.Index))
		block.Index = id
	} else {
		return nil, fmt.Errorf("index is empty")
	}

	return &block, nil
}

func (b *Block) toGormBlock() *gormBlock {
	return &gormBlock{
		ID:       b.ID,
		Type:     b.Type,
		ParentID: b.ParentID,
		Index:    string(b.Index.Bytes()),
		Deleted:  false,
		Erased:   false,
	}
}
