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

func TestInsertOp(t *testing.T) {
	var err error
	store := NewMemStore()
	space := &Space{
		ID:   sid,
		Name: "",
	}
	err = store.CreateSpace(space)
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
	err = store.ApplyChange(&space.ID, changes)
	assert.NoError(t, err)

	block, err := store.GetBlock(&space.ID, b1)
	assert.NoError(t, err)
	assert.Equal(t, b1, block.ID)

	store.Print(&sid)
}

func TestMoveOp(t *testing.T) {
}
