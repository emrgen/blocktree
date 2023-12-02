package blocktree

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var (
	_ Store = (*GormStore)(nil)
)

// GormSpace is a space in gorm database.
type GormSpace struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key"`
	Name string    `gorm:"not null"`
}

func (s *GormSpace) toSpace() *Space {
	//TODO implement me
	panic("implement me")
}

func (s *Space) toGormSpace() *GormSpace {
	//TODO implement me
	panic("implement me")
}

type GormBlock struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key"`
	Type string    `gorm:"not null"`
	// nullable
	ParentID uuid.UUID `gorm:"type:uuid;not null"`
	Index    string    `gorm:"not null"`
	Deleted  bool      `gorm:"not null"`
	Erased   bool      `gorm:"not null"`
}

func (b *GormBlock) toBlock() (*Block, error) {
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

func (b *Block) toGormBlock() *GormBlock {
	return &GormBlock{
		ID:       b.ID,
		Type:     b.Type,
		ParentID: b.ParentID,
		Index:    string(b.Index.Bytes()),
		Deleted:  false,
		Erased:   false,
	}
}

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (g GormStore) CreateSpace(space *Space) error {
	model := space.toGormSpace()
	return g.db.Create(model).Error
}

func (g GormStore) CreateBlock(spaceID *SpaceID, block *Block) error {
	model := block.toGormBlock()
	return g.db.Create(model).Error
}

func (g GormStore) GetBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	var model GormBlock
	res := g.db.First(&model, "id = ?", id)
	if res.Error != nil {
		return nil, res.Error
	}

	if res.RecordNotFound() {
		return nil, gorm.ErrRecordNotFound
	}
	return model.toBlock()
}

func (g GormStore) GetChildrenBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	var blocks []*GormBlock
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

func (g GormStore) GetDescendantBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetParentBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetBlocks(spaceID *SpaceID, ids []BlockID) ([]*Block, error) {
	var blocks []*GormBlock
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

func (g GormStore) GetAncestorEdges(spaceID *SpaceID, id []BlockID) ([]BlockEdge, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) GetTransaction(spaceID *SpaceID, id *TransactionID) (*Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) PutTransactions(spaceID *SpaceID, tx []*Transaction) error {
	//TODO implement me
	panic("implement me")
}

func (g GormStore) ApplyChange(space *SpaceID, change *StoreChange) error {
	//TODO implement me
	panic("implement me")
}
