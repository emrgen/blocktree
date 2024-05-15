package blocktree

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormStore_Create(t *testing.T) {
	err := os.MkdirAll("./tmp", os.ModePerm)
	assert.NoError(t, err)

	db, err := gorm.Open(sqlite.Open("./tmp/blocktree.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	store := NewGormStore(db)
	_ = store.Migrate()

	s := newSpace(s1, "physics")
	err = store.CreateSpace(s)
	assert.NoError(t, err)

	block, err := store.GetBlock(&s1, s1)
	assert.NoError(t, err)
	logrus.Print(block)
	assert.Equal(t, block.ID, s1)

	_ = os.Remove("./tmp/blocktree.db")
}

func TestGormStore_CreateSpace(t *testing.T) {

}

func TestGormStore_CreateBlock(t *testing.T) {

}
