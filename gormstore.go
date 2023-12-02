package blocktree

import (
	"github.com/jinzhu/gorm"
)

var (
	_ Store = (*GormStore)(nil)
)

type spaceModel struct {
	ID   string `gorm:"primary_key"`
	Name string `gorm:"not null"`
}

func (s *spaceModel) toSpace() *Space {
	//TODO implement me
	panic("implement me")
}

func (s *Space) toSpaceModel() *spaceModel {
	//TODO implement me
	panic("implement me")
}

type blockModel struct {
	ID       string `gorm:"primary_key"`
	Type     string `gorm:"not null"`
	ParentID string `gorm:"not null"`
	Index    string `gorm:"not null"`
	Deleted  bool   `gorm:"not null"`
	Erased   bool   `gorm:"not null"`
}

func (b *blockModel) toBlock() *Block {
	//TODO implement me
	panic("implement me")
}

func (b *Block) toBlockModel() *blockModel {
	//TODO implement me
	panic("implement me")
}

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (g GormStore) CreateSpace(space *Space) error {
	model := space.toSpaceModel()
	return g.db.Create(model).Error
}

func (g GormStore) CreateBlock(spaceID *SpaceID, block *Block) error {
	model := block.toBlockModel()
	return g.db.Create(model).Error
}

func (g GormStore) GetBlock(spaceID *SpaceID, id BlockID) (*Block, error) {
	var model blockModel
	res := g.db.First(&model, "id = ?", id)
	if res.Error != nil {
		return nil, res.Error
	}

	if res.RecordNotFound() {
		return nil, gorm.ErrRecordNotFound
	}
	return model.toBlock(), nil
}

func (g GormStore) GetChildrenBlocks(spaceID *SpaceID, id BlockID) ([]*Block, error) {
	var blocks []*blockModel
	res := g.db.Where("parent_id = ?", id).Find(&blocks)
	if res.Error != nil {
		return nil, res.Error
	}

	ret := make([]*Block, len(blocks))
	for i, b := range blocks {
		ret[i] = b.toBlock()
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
	var blocks []*blockModel
	res := g.db.Where("id IN (?)", ids).Find(&blocks)
	if res.Error != nil {
		return nil, res.Error
	}

	ret := make([]*Block, len(blocks))
	for i, b := range blocks {
		ret[i] = b.toBlock()
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
