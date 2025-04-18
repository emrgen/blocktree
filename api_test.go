package blocktree

import (
	"fmt"
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

	tx = createTx(s1, moveOp(b2, s1, s1, PositionStart))
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

	tx = createTx(s1, moveOp(b1, s1, s1, PositionEnd))
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

	tx = createTx(s1, moveOp(b1, s1, b2, PositionAfter))
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

	tx = createTx(s1, moveOp(b2, s1, b1, PositionBefore))
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

	tx = createTx(s1, moveOp(b1, s1, b3, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetChildrenBlocks(s1, b3)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(blocks))

	// should be a cycle
	tx = createTx(s1, moveOp(b1, s1, b1, PositionStart))
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
	block, err := api.GetBlock(s1, b1)
	assert.NoError(t, err)

	assert.Equal(t, b1, block.ID)
	blocks, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blocks))
}

func TestApi_TransactionChanges(t *testing.T) {
	var err error

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx1 := createTx(s1, insertOp(b1, "p1", s1, PositionEnd))
	ch1, err := api.Apply(tx1)
	assert.NoError(t, err)

	tx2 := createTx(s1, insertOp(b2, "p2", s1, PositionEnd))
	ch2, err := api.Apply(tx2)
	assert.NoError(t, err)

	tx3 := createTx(s1, insertOp(b3, "p3", s1, PositionEnd))
	ch3, err := api.Apply(tx3)
	assert.NoError(t, err)

	tx4 := createTx(s1, moveOp(b1, s1, b3, PositionStart))
	ch4, err := api.Apply(tx4)
	assert.NoError(t, err)

	changes := tx1.changes
	fmt.Println(changes)

	assert.Equal(t, 1, ch1.children.Size())
	assert.Equal(t, 1, ch2.children.Size())
	assert.Equal(t, 1, ch3.children.Size())

	//assert.Equal(t, 1, ch4.updated.Size())
	assert.Equal(t, 2, ch4.children.Size())

	transactions, err := api.store.GetNextTransactions(&s1, tx1.ID, 0, 10)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(transactions))

	assert.Equal(t, tx2.ID, transactions[0].ID)

	assert.Equal(t, 1, ch1.inserted.Size())
	assert.Equal(t, ch2, transactions[0].changes)
	assert.Equal(t, ch3, transactions[1].changes)
	assert.Equal(t, ch4, transactions[2].changes)

	updates, err := api.GetUpdates(s1, tx1.ID)
	assert.NoError(t, err)

	assert.Equal(t, b1, updates.Children[b3][0])
	assert.Equal(t, []BlockID{b2, b3}, updates.Children[s1])

	assert.Equal(t, 3, len(updates.Blocks))

	//fmt.Printf("updates: %v\n", updates)
}

func TestApi_AddBackLink(t *testing.T) {
	var err error
	var tx *Transaction

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b1, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, insertOp(b2, "p1", s1, PositionStart))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	tx = createTx(s1, linkInsertOp(b3, "p1", b1))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err := api.GetLinkedBlocks(s1, b1)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(blocks))

	tx = createTx(s1, linkOp(b3, b2))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err = api.GetLinkedBlocks(s1, b2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blocks))

	links, err := api.GetBackLinks(s1, b3)
	assert.NoError(t, err)
	assert.Equal(t, len(links), 1)

	blocks, err = api.GetLinkedBlocks(s1, b1)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(blocks))

	tx = createTx(s1, unlinkOp(b3))
	_, err = api.Apply(tx)
	assert.NoError(t, err)

	blocks, err = api.GetLinkedBlocks(s1, b2)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(blocks))

	links, err = api.GetBackLinks(s1, b3)
	assert.NoError(t, err)
	assert.Equal(t, len(links), 0)
}

func TestApi_IdempotentTransaction(t *testing.T) {
	var err error

	api := NewApi(NewMemStore())

	err = api.CreateSpace(s1, "test-1")
	assert.NoError(t, err)

	tx1 := createTx(s1, insertOp(b1, "p1", s1, PositionEnd))
	tx2 := createTx(s1, insertOp(b2, "p1", s1, PositionEnd))
	_, err = api.Apply(tx1)
	assert.NoError(t, err)

	_, err = api.Apply(tx1, tx2)
	assert.NoError(t, err)

	after, err := api.GetChildrenBlocks(s1, s1)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(after))

	//api.store.(*MemStore).Print(&s1)
}
