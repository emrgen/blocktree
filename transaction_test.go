package blocktree

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	s1, _ = uuid.Parse("00000000-0000-0000-0001-000000000001")
	s2, _ = uuid.Parse("00000000-0000-0000-0001-000000000002")
	//s3, _ = uuid.Parse("00000000-0000-0000-0001-000000000003")
	b1, _ = uuid.Parse("00000000-0000-0000-0000-000000000001")
	b2, _ = uuid.Parse("00000000-0000-0000-0000-000000000002")
	b3, _ = uuid.Parse("00000000-0000-0000-0000-000000000003")
	b4, _ = uuid.Parse("00000000-0000-0000-0000-000000000004")
	b5, _ = uuid.Parse("00000000-0000-0000-0000-000000000005")
	b6, _ = uuid.Parse("00000000-0000-0000-0000-000000000006")
	b7, _ = uuid.Parse("00000000-0000-0000-0000-000000000007")
	b8, _ = uuid.Parse("00000000-0000-0000-0000-000000000008")
	b9, _ = uuid.Parse("00000000-0000-0000-0000-000000000009")
	//b10, _ = uuid.Parse("00000000-0000-0000-0000-000000000010")
)

func insertOp(blockID uuid.UUID, object string, refID uuid.UUID, pos PointerPosition) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeInsert,
		Object:  object,
		BlockID: blockID,
		At: &Pointer{
			BlockID:  refID,
			Position: pos,
		},
	}
}

func deleteOp(blockID uuid.UUID) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeDelete,
		BlockID: blockID,
	}
}

func eraseOp(blockID uuid.UUID) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeErase,
		BlockID: blockID,
	}
}

func moveOp(blockID uuid.UUID, from, refID uuid.UUID, pos PointerPosition) Op {
	return Op{
		Table:    "block",
		Type:     OpTypeMove,
		BlockID:  blockID,
		ParentID: &from,
		At: &Pointer{
			BlockID:  refID,
			Position: pos,
		},
	}
}

func updateOp(blockID uuid.UUID, props []byte) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeUpdate,
		BlockID: blockID,
		Props:   props,
	}
}

func linkInsertOp(blockID uuid.UUID, object string, refID uuid.UUID) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeInsert,
		Object:  object,
		Linked:  true,
		BlockID: blockID,
		At: &Pointer{
			BlockID:  refID,
			Position: PositionInside,
		},
	}
}

func linkOp(blockID uuid.UUID, refID uuid.UUID) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeLink,
		BlockID: blockID,
		At: &Pointer{
			BlockID:  refID,
			Position: PositionInside,
		},
	}
}

func unlinkOp(blockID uuid.UUID) Op {
	return Op{
		Table:   "block",
		Type:    OpTypeUnlink,
		BlockID: blockID,
	}
}

func patchOp(blockID uuid.UUID, patch []byte) Op {
	return Op{
		Table:   "block",
		Type:    OpTypePatch,
		BlockID: blockID,
		Patch:   patch,
	}
}

func createTx(spaceID uuid.UUID, ops ...Op) *Transaction {
	return &Transaction{
		ID:      uuid.New(),
		SpaceID: spaceID,
		UserID:  uuid.Nil,
		Ops:     ops,
	}
}

func createSpace(store *MemStore, spaceID uuid.UUID) error {
	space := &Space{
		ID:   spaceID,
		Name: "test-space",
	}
	return store.CreateSpace(space)
}

func TestInsertOp(t *testing.T) {
	var err error
	store := NewMemStore()
	err = createSpace(store, s1)
	assert.NoError(t, err)

	// create a block transaction
	tx := createTx(s1,
		insertOp(b1, "page", s1, PositionStart),
		insertOp(b2, "title", b1, PositionStart),
		insertOp(b3, "p1", b2, PositionAfter),
		insertOp(b4, "p2", b3, PositionBefore),
		insertOp(b5, "p3", b1, PositionEnd),
	)

	// apply the transaction
	changes, err := tx.prepare(store)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		return changes != nil
	})

	// apply the change to the store
	err = store.Apply(tx, changes)
	assert.NoError(t, err)

	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, b1, block.ID)

	//store.Print(&s1)
}

func TestInsertOpBetween(t *testing.T) {
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
		},
	}

	applyTransaction(t, store, tx)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b3, "p1", b2, PositionAfter),
			insertOp(b4, "p2", b3, PositionBefore),
			insertOp(b5, "p3", b1, PositionEnd),
		},
	}

	applyTransaction(t, store, tx)

	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, b1, block.ID)

	//store.Print(&s1)
}

func TestInsertAfterOp(t *testing.T) {
	var err error
	var tx *Transaction
	store := NewMemStore()
	err = createSpace(store, s1)
	assert.NoError(t, err)

	// create a block transaction
	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b1, "p1", s1, PositionEnd),
		},
	}
	applyTransaction(t, store, tx)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b2, "p2", s1, PositionEnd),
		},
	}
	applyTransaction(t, store, tx)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b3, "p3", b1, PositionAfter),
		},
	}
	applyTransaction(t, store, tx)

	//store.Print(&s1)
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

	changes, _ := tx.prepare(store)
	err = store.Apply(tx, changes)
	return err
}

func applyTransaction(t *testing.T, store *MemStore, tx *Transaction) {
	changes, err := tx.prepare(store)
	assert.NoError(t, err)

	assert.Condition(t, func() bool {
		return changes != nil
	})

	err = store.Apply(tx, changes)
	assert.NoError(t, err)
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
			moveOp(b1, s1, b2, PositionStart),
			moveOp(b3, s1, b2, PositionEnd),
		},
	}

	_, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)

	applyTransaction(t, store, tx)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			moveOp(b4, s1, b1, PositionAfter),
			moveOp(b5, s1, b3, PositionBefore),
		},
	}

	applyTransaction(t, store, tx)

	blocks, err := store.GetChildrenBlocks(&s1, b2)
	assert.NoError(t, err)

	assert.Equal(t, 4, len(blocks), "b1 should have 4 children blocks")
	assert.Equal(t, b1, blocks[0].ID)
	assert.Equal(t, b4, blocks[1].ID)
	assert.Equal(t, b5, blocks[2].ID)
	assert.Equal(t, b3, blocks[3].ID)

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
			moveOp(b1, s1, b2, PositionStart),
			moveOp(b2, s1, b1, PositionStart),
		},
	}

	_, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)

	_, err = tx.prepare(store)
	assert.EqualError(t, err, ErrCreatesCycle.Error())
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
			moveOp(b1, s1, b2, PositionStart),
			moveOp(b2, s1, b3, PositionStart),
			moveOp(b3, s1, b4, PositionStart),
			moveOp(b4, s1, b1, PositionStart),
		},
	}

	_, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)

	_, err = tx.prepare(store)
	assert.EqualError(t, err, ErrCreatesCycle.Error())
}

func TestBlockLink(t *testing.T) {
	var err error
	store := NewMemStore()
	err = prepareSpace(store, s1)
	assert.NoError(t, err)

	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			linkInsertOp(b6, "l1", b2),
			linkInsertOp(b7, "l2", b2),
			insertOp(b8, "p2", b2, PositionEnd),
			insertOp(b9, "p2", b6, PositionEnd),
		},
	}

	change, err := tx.prepare(store)
	assert.NoError(t, err)

	err = store.Apply(tx, change)
	assert.NoError(t, err)

	space, err := store.getSpace(&s1)
	assert.NoError(t, err)

	assert.Equal(t, 3, space.children[b2].Len(), "b2 should have total 3 linked+child blocks")

	blocks, err := store.GetChildrenBlocks(&s1, b2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blocks), "b2 should have 1 children block")

	descendants, err := store.GetDescendantBlocks(&s1, s1)
	assert.NoError(t, err)
	v1, err := blockViewFromBlocks(s1, descendants)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		return len(v1.Children) == 5
	})

	//v1.Print()
}

func TestPatchOp(t *testing.T) {
	var err error
	var tx *Transaction
	store := NewMemStore()
	err = store.CreateSpace(&Space{
		ID:   s1,
		Name: "s1",
	})
	assert.NoError(t, err)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b1, "p1", s1, PositionEnd),
			insertOp(b2, "p2", s1, PositionEnd),
		},
	}
	applyTransaction(t, store, tx)

	patch := []byte(`[{"op":"add","path":"/name","value":"John Doe"}]`)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			patchOp(b1, patch),
		},
	}
	applyTransaction(t, store, tx)

	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"John Doe"}`, string(block.Json.Content))

	patch = []byte(`[{"op":"add","path":"/age","value":30}]`)
	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			patchOp(b1, patch),
		},
	}
	applyTransaction(t, store, tx)

	block, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"John Doe","age":30}`, string(block.Json.Content))

	//store.Print(&s1)
}

func TestUpdateBlockProps(t *testing.T) {
	var err error
	var tx *Transaction
	store := NewMemStore()
	err = store.CreateSpace(&Space{
		ID:   s1,
		Name: "s1",
	})
	assert.NoError(t, err)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			insertOp(b1, "p1", s1, PositionEnd),
		},
	}
	applyTransaction(t, store, tx)

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			updateOp(b1, []byte(`[{"op":"add","path":"/name","value":"John Doe"}, {"op":"add","path":"/age","value":30}]`)),
		},
	}
	applyTransaction(t, store, tx)

	block, err := store.GetBlock(&s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"John Doe","age":30}`, block.Props.String())

	tx = &Transaction{
		ID:      uuid.New(),
		SpaceID: s1,
		Ops: []Op{
			updateOp(b1, []byte(`[{"op":"replace","path":"/name","value":"Jane Doe"}]`)),
			updateOp(b1, []byte(`[{"op":"add","path":"/age","value":20},{"op":"add","path":"/pin","value":700010}]`)),
		},
	}
	applyTransaction(t, store, tx)

	block, err = store.GetBlock(&s1, b1)
	assert.NoError(t, err)

	assert.Equal(t, `{"name":"Jane Doe","age":20,"pin":700010}`, block.Props.String())
}

func TestInsertOpSyncBlocks(t *testing.T) {
	var err error
	store := NewMemStore()
	err = createSpace(store, s1)
	assert.NoError(t, err)

	tx := createTx(s1,
		insertOp(b1, "page", s1, PositionStart),
		insertOp(b2, "title", b1, PositionStart),
		insertOp(b3, "p1", b2, PositionAfter),
		insertOp(b4, "p2", b3, PositionBefore),
	)

	changes, err := tx.prepare(store)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		return changes != nil
	})

	sb := changes.intoSyncBlocks()
	assert.Equal(t, 2, sb.children.Size())
}
