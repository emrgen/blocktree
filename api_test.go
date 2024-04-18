package blocktree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSpace(t *testing.T) {
	api := NewApi(NewMemStore())

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
	api := NewApi(NewMemStore())

	err := api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	err = api.CreateSpace(s1, "test-1")
	assert.Error(t, err)
}

func TestInsertBlockAtEnd(t *testing.T) {
	var err error
	var tx *Transaction

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
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

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionStart))
	_, err = api.Apply(tx)
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

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b3, "p3", b1, PositionAfter))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b4, "p3", b2, PositionBefore))
	sb, err := api.Apply(tx)
	assert.NoError(t, err)
	assert.Equal(t, 1, sb.children.Size())

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 4, len(blocks))
	assert.Equal(t, b1, blocks[0].ID)
	assert.Equal(t, b3, blocks[1].ID)
	assert.Equal(t, b4, blocks[2].ID)
	assert.Equal(t, b2, blocks[3].ID)
}

func TestDeleteBlock(t *testing.T) {
	var err error
	var tx *Transaction

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, deleteOp(b1))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	type blockState struct {
		ID      BlockID
		Deleted bool
	}

	ids := make([]blockState, 0)
	for _, b := range blocks {
		ids = append(ids, blockState{
			ID:      b.ID,
			Deleted: b.Deleted,
		})
	}

	assert.Equal(t, []blockState{{ID: b1, Deleted: true}, {ID: b2, Deleted: false}}, ids)
}

func TestEraseBlock(t *testing.T) {
	var err error
	var tx *Transaction

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, eraseOp(b1))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	type blockState struct {
		ID     BlockID
		Erased bool
	}

	ids := make([]blockState, 0)
	for _, b := range blocks {
		ids = append(ids, blockState{
			ID:     b.ID,
			Erased: b.Erased,
		})
	}

	assert.Equal(t, []blockState{{ID: b1, Erased: true}, {ID: b2, Erased: false}}, ids)

	block, err := api.GetBlock(s1, b1)
	assert.NoError(t, err)
	assert.True(t, block.Erased)
}

func TestMoveBlockStart(t *testing.T) {
	var err error
	var tx *Transaction

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b2, s1, PositionStart))
	_, err = api.Apply(tx)
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

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b1, s1, PositionEnd))
	_, err = api.Apply(tx)
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

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b1, b2, PositionAfter))
	_, err = api.Apply(tx)
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

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b2, b1, PositionBefore))
	_, err = api.Apply(tx)
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

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	// b1 -> b2 -> b3 -> b1
	tx = createTx(s1, insertOp(b1, "p1", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p2", b1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b3, "p3", b2, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, moveOp(b1, b3, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, b3)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(blocks))

	// should be a cycle
	tx = createTx(s1, moveOp(b1, b1, PositionStart))
	_, err = api.Apply(tx)
	assert.Error(t, err)

	blocks, err = api.GetChildrenBlocks(s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blocks))
}

func TestApi_GetBlockSpaceID(t *testing.T) {
	var err error
	var tx *Transaction

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionEnd))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	spaceID, err := api.GetBlockSpaceID(b1)
	assert.NoError(t, err)
	assert.Equal(t, s1, *spaceID)
}
