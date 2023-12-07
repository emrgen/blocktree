package blocktree

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSpace(t *testing.T) {
	api := New(NewMemStore())

	err := api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	err = api.CreateSpace(s2, "test-2")
	assert.NoError(t, err)

	block, err := api.GetBlock(s1, s1)
	assert.NoError(t, err)
	assert.Equal(t, s1, block.ID)

	block, err = api.GetBlock(s2, s2)
	assert.NoError(t, err)
	assert.Equal(t, s2, block.ID)
}

func TestCreateSpaceAlreadyExists(t *testing.T) {
	api := New(NewMemStore())

	err := api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	err = api.CreateSpace(s1, "test-1")
	assert.Error(t, err)
}

func TestInsertBlockAtEnd(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, b1, blocks[0].ID)
	assert.Equal(t, b2, blocks[1].ID)
}

func TestInsertBlockAtStart(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, b2, blocks[0].ID)
	assert.Equal(t, b1, blocks[1].ID)
}

func TestInsertBlockAfter(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b3, "p3", b1, PositionAfter))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b4, "p3", b2, PositionBefore))
	err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 4, len(blocks))
	assert.Equal(t, b1, blocks[0].ID)
	assert.Equal(t, b3, blocks[1].ID)
	assert.Equal(t, b4, blocks[2].ID)
	assert.Equal(t, b2, blocks[3].ID)
}

func TestMoveBlockStart(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b2, s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, b2, blocks[0].ID)
	assert.Equal(t, b1, blocks[1].ID)
}

func TestMoveBlockEnd(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b1, s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, b2, blocks[0].ID)
	assert.Equal(t, b1, blocks[1].ID)
}

func TestMoveBlockAfter(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b1, b2, PositionAfter))
	err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, b2, blocks[0].ID)
	assert.Equal(t, b1, blocks[1].ID)
}

func TestMoveBlockBefore(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b2, b1, PositionBefore))
	err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, b2, blocks[0].ID)
	assert.Equal(t, b1, blocks[1].ID)
}

func TestSimpleMoveCycle(t *testing.T) {
	var err error
	var tx *Transaction

	api := New(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	// b1 -> b2 -> b3 -> b1
	tx = createTx(s1, insertOp(b1, "p1", s1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", b1, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b3, "p3", b2, PositionEnd))
	err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b1, b3, PositionStart))
	err = api.Apply(tx)

	blocks, err := api.GetChildrenBlocks(s1, b3)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(blocks))

	tx = createTx(s1, moveOp(b1, b1, PositionStart))
	err = api.Apply(tx)

	blocks, err = api.GetChildrenBlocks(s1, b1)
	assert.Equal(t, 1, len(blocks))
}
