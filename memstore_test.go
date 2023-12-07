package blocktree

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsertBlockIntoSpace(t *testing.T) {
	store := NewMemStore()
	s := NewSpace(s1, "physics")
	err := store.CreateSpace(s)
	assert.NoError(t, err)

	bl1 := NewBlock(b1, s1, "p1")
	err = store.CreateBlock(&s1, bl1)
	assert.NoError(t, err)
	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, bl1, block)
}

func TestInsertMultipleBlocks(t *testing.T) {
	store := NewMemStore()
	s := &Space{
		ID:   s1,
		Name: "physics",
	}
	err := store.CreateSpace(s)
	assert.NoError(t, err)

	bl1 := NewBlock(b1, s1, "p1")
	err = store.CreateBlock(&s1, bl1)
	assert.NoError(t, err)
	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, bl1, block)

	bl2 := NewBlock(b2, b1, "p2")
	err = store.CreateBlock(&s1, bl2)
	assert.NoError(t, err)
	block, err = store.GetBlock(&s1, b2)
	assert.NoError(t, err)
	assert.Equal(t, bl2, block)

	bl3 := NewBlock(b3, b2, "page")
	err = store.CreateBlock(&s1, bl3)
	assert.NoError(t, err)
	block, err = store.GetBlock(&s1, b3)
	assert.NoError(t, err)
	assert.Equal(t, bl3, block)

	bl4 := NewBlock(b4, b3, "p4")
	err = store.CreateBlock(&s1, bl4)
	assert.NoError(t, err)
	block, err = store.GetBlock(&s1, b4)
	assert.NoError(t, err)
	assert.Equal(t, bl4, block)

	// check the parent-child relationship
	blocks, err := store.GetChildrenBlocks(&s1, s1)
	assert.NoError(t, err)
	assert.Equal(t, []*Block{bl1}, blocks)

	// check the parent-descendant relationship
	blocks, err = store.GetDescendantBlocks(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(blocks))
	assert.Equal(t, []*Block{bl1, bl2, bl3}, blocks)
}

func TestInsertMultipleBlocksInMultipleSpaces(t *testing.T) {
	sp1 := NewSpace(s1, "physics")
	sp2 := NewSpace(s2, "chemistry")
	store := NewMemStore()
	err := store.CreateSpace(sp1)
	assert.NoError(t, err)
	err = store.CreateSpace(sp2)
	assert.NoError(t, err)

	bl1 := NewBlock(b1, s1, "p1")
	err = store.CreateBlock(&s1, bl1)
	assert.NoError(t, err)
	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, bl1, block)

	bl2 := NewBlock(b2, s2, "p2")
	err = store.CreateBlock(&s2, bl2)
	assert.NoError(t, err)
	block, err = store.GetBlock(&s2, b2)
	assert.NoError(t, err)
	assert.Equal(t, bl2, block)

	blocks, err := store.GetChildrenBlocks(&s1, s1)
	assert.NoError(t, err)
	assert.Equal(t, []*Block{bl1}, blocks)

	blocks, err = store.GetChildrenBlocks(&s2, s2)
	assert.NoError(t, err)
	logrus.Info(blocks)
	assert.Equal(t, []*Block{bl2}, blocks)

	store.Print(&s1)
}
