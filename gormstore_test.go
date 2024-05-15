package blocktree

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createStore() (Store, error) {
	err := os.MkdirAll("./tmp", os.ModePerm)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open("./tmp/blocktree.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	store := NewGormStore(db)
	_ = store.Migrate()

	return store, nil
}

func cleanDB() {
	_ = os.Remove("./tmp/blocktree.db")
}

func TestGormStore_CreateSpace(t *testing.T) {
	store, err := createStore()
	assert.NoError(t, err)
	defer cleanDB()

	s := newSpace(s1, "physics")
	err = store.CreateSpace(s)
	assert.NoError(t, err)

	block, err := store.GetBlock(&s1, s1)
	assert.NoError(t, err)
	logrus.Print(block)
	assert.Equal(t, block.ID, s1)

}

func TestGormStore_CreateBlock(t *testing.T) {
	var err error

	store, err := createStore()
	assert.NoError(t, err)
	defer cleanDB()

	s := newSpace(s1, "physics")
	err = store.CreateSpace(s)
	assert.NoError(t, err)

	err = store.CreateBlock(&s1, &Block{
		Type:     "p1",
		Table:    "block",
		ID:       b1,
		ParentID: s1,
		Index:    DefaultFracIndex(),
	})
	assert.NoError(t, err)

	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, block.ID, b1)
}
