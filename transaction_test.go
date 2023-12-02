package blocktree

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	s1, _  = uuid.Parse("00000000-0000-0000-0001-000000000001")
	s2, _  = uuid.Parse("00000000-0000-0000-0001-000000000002")
	s3, _  = uuid.Parse("00000000-0000-0000-0001-000000000003")
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
	err = createSpace(store, s1)
	assert.NoError(t, err)

	// create a block transaction
	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b1, "page", s1, PositionStart),
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
	err = store.ApplyChange(&s1, changes)
	assert.NoError(t, err)

	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, b1, block.ID)

	store.Print(&s1)
}

func prepareSpace(store *MemStore, spaceID uuid.UUID) error {
	var err error
	err = createSpace(store, s1)
	if err != nil {
		return err
	}

	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b1, "p1", s1, PositionEnd),
			insertOp(b2, "p2", s1, PositionEnd),
			insertOp(b3, "p3", s1, PositionEnd),
			insertOp(b4, "p4", s1, PositionEnd),
			insertOp(b5, "p5", s1, PositionEnd),
		},
	}

	changes, _ := tx.Prepare(store)
	err = store.ApplyChange(&s1, changes)
	return err
}

func TestMoveOp(t *testing.T) {
	var err error
	store := NewMemStore()
	err = prepareSpace(store, s1)
	assert.NoError(t, err)

	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			moveOp(b1, b2, PositionStart),
			moveOp(b3, b2, PositionEnd),
			moveOp(b4, b1, PositionAfter),
			moveOp(b5, b3, PositionBefore),
		},
	}

	_, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)

	changes, err := tx.Prepare(store)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		return changes != nil
	})

	err = store.ApplyChange(&s1, changes)
	assert.NoError(t, err)

	store.Print(&s1)
}

func TestMoveOpWithSimpleCycle(t *testing.T) {
	var err error
	store := NewMemStore()
	err = prepareSpace(store, s1)
	assert.NoError(t, err)

	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			moveOp(b1, b2, PositionStart),
			moveOp(b2, b1, PositionStart),
		},
	}

	_, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)

	_, err = tx.Prepare(store)
	assert.EqualError(t, err, ErrDetectedCycle.Error())
}

func TestMoveOpWithComplexCycle(t *testing.T) {
	var err error
	store := NewMemStore()
	err = prepareSpace(store, s1)
	assert.NoError(t, err)

	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			moveOp(b1, b2, PositionStart),
			moveOp(b2, b3, PositionStart),
			moveOp(b3, b4, PositionStart),
			moveOp(b4, b1, PositionStart),
		},
	}

	_, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)

	_, err = tx.Prepare(store)
	assert.EqualError(t, err, ErrDetectedCycle.Error())
}
