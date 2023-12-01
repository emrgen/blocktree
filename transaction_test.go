package blocktree

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	sid, _ = uuid.Parse("00000000-0000-0000-0000-000000000000")
	b1, _  = uuid.Parse("00000000-0000-0000-0000-000000000001")
	b2, _  = uuid.Parse("00000000-0000-0000-0000-000000000002")
	b3, _  = uuid.Parse("00000000-0000-0000-0000-000000000003")
	b4, _  = uuid.Parse("00000000-0000-0000-0000-000000000004")
	b5, _  = uuid.Parse("00000000-0000-0000-0000-000000000005")
	b6, _  = uuid.Parse("00000000-0000-0000-0000-000000000006")
	b7, _  = uuid.Parse("00000000-0000-0000-0000-000000000007")
	b8, _  = uuid.Parse("00000000-0000-0000-0000-000000000008")
	b9, _  = uuid.Parse("00000000-0000-0000-0000-000000000009")
	b10, _ = uuid.Parse("00000000-0000-0000-0000-000000000010")
)

func insertOp(blockID uuid.UUID, typ string, refID uuid.UUID, pos PointerPosition) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeInsert,
		BlockID: blockID,
		At: &Pointer{
			BlockID:  refID,
			Position: pos,
		},
		Props: map[string]interface{}{
			"type": typ,
		},
	}
}

func moveOp(blockID uuid.UUID, refID uuid.UUID, pos PointerPosition) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeMove,
		BlockID: blockID,
		At: &Pointer{
			BlockID:  refID,
			Position: pos,
		},
	}
}

func createSpace(store *MemStore, spaceID uuid.UUID) error {
	space := &Space{
		ID:   spaceID,
		Name: "",
	}
	return store.CreateSpace(space)
}

func TestInsertOp(t *testing.T) {
	var err error
	store := NewMemStore()
	err = createSpace(store, sid)
	assert.NoError(t, err)

	// create a block transaction
	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: sid,
		Ops: []Op{
			insertOp(b1, "page", sid, PositionStart),
			insertOp(b2, "title", b1, PositionStart),
			insertOp(b3, "p1", b2, PositionAfter),
			insertOp(b4, "p2", b3, PositionBefore),
			insertOp(b5, "p3", b1, PositionEnd),
		},
	}

	// apply the transaction
	changes, err := tx.Prepare(store)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		return changes != nil
	})

	// apply the change to the store
	err = store.ApplyChange(&sid, changes)
	assert.NoError(t, err)

	block, err := store.GetBlock(&sid, b1)
	assert.NoError(t, err)
	assert.Equal(t, b1, block.ID)

	store.Print(&sid)
}

func prepareSpace(store *MemStore, spaceID uuid.UUID) error {
	var err error
	err = createSpace(store, sid)
	if err != nil {
		return err
	}

	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: sid,
		Ops: []Op{
			insertOp(b1, "p1", sid, PositionEnd),
			insertOp(b2, "p2", sid, PositionEnd),
			insertOp(b3, "p3", sid, PositionEnd),
			insertOp(b4, "p4", sid, PositionEnd),
			insertOp(b5, "p5", sid, PositionEnd),
		},
	}

	changes, _ := tx.Prepare(store)
	err = store.ApplyChange(&sid, changes)
	return err
}

func TestMoveOp(t *testing.T) {
	var err error
	store := NewMemStore()
	err = prepareSpace(store, sid)
	assert.NoError(t, err)

	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: sid,
		Ops: []Op{
			moveOp(b1, b2, PositionStart),
			moveOp(b3, b2, PositionEnd),
			moveOp(b4, b1, PositionAfter),
			moveOp(b5, b3, PositionBefore),
		},
	}

	_, err = store.GetBlock(&sid, b1)
	assert.NoError(t, err)

	changes, err := tx.Prepare(store)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		return changes != nil
	})

	err = store.ApplyChange(&sid, changes)
	assert.NoError(t, err)

	store.Print(&sid)
}
