package blocktree

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsertBlockIntoSpace(t *testing.T) {
	store := NewMemStore()
	s := &Space{
		ID:   sid,
		Name: "physics",
	}
	err := store.CreateSpace(s)
	assert.NoError(t, err)

	bl1 := NewBlock(b1, &sid, "p1")
	err = store.CreateBlock(&sid, bl1)
	assert.NoError(t, err)
	block, err := store.GetBlock(&sid, b1)
	assert.NoError(t, err)
	assert.Equal(t, bl1, block)
}

func TestInsertMultipleBlocks(t *testing.T) {
	store := NewMemStore()
	s := &Space{
		ID:   sid,
		Name: "physics",
	}
	err := store.CreateSpace(s)
	assert.NoError(t, err)

	bl1 := NewBlock(b1, &sid, "p1")
	err = store.CreateBlock(&sid, bl1)
	assert.NoError(t, err)
	block, err := store.GetBlock(&sid, b1)
	assert.NoError(t, err)
	assert.Equal(t, bl1, block)

	bl2 := NewBlock(b2, &b1, "p2")
	err = store.CreateBlock(&sid, bl2)
	assert.NoError(t, err)
	block, err = store.GetBlock(&sid, b2)
	assert.NoError(t, err)
	assert.Equal(t, bl2, block)

	bl3 := NewBlock(b3, &b2, "p3")
	err = store.CreateBlock(&sid, bl3)
	assert.NoError(t, err)
	block, err = store.GetBlock(&sid, b3)
	assert.NoError(t, err)
	assert.Equal(t, bl3, block)

	blocks, err := store.GetChildrenBlocks(&sid, sid)
	assert.NoError(t, err)
	assert.Equal(t, []*Block{bl1}, blocks)

	blocks, err = store.GetDescendantBlocks(&sid, b1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(blocks))
	assert.Equal(t, []*Block{bl1, bl2, bl3}, blocks)
}
