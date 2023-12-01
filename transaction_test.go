package blocktree

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemSpace(t *testing.T) {
	var err error
	store := NewMemStore()
	space := &Space{
		ID:   uuid.New(),
		Name: "",
	}
	err = store.CreateSpace(space)
	assert.NoError(t, err)

	// create a block transaction
	b1 := uuid.New()
	tx := &Transaction{
		ID:      uuid.New(),
		SpaceID: space.ID,
		Ops: []Op{
			{
				Table:   "block",
				Type:    OpTypeInsert,
				BlockID: b1,
				At: &Pointer{
					BlockID:  space.ID,
					Position: PositionStart,
				},
				Props: map[string]interface{}{
					"type": "page",
				},
			},
		},
	}

	// apply the transaction
	changes, err := tx.Prepare(store)
	assert.NoError(t, err)
	logrus.Info(changes.blockChange.inserted.Cardinality())
	changes.blockChange.inserted.ForEach(func(item *Block) bool {
		logrus.Infof("inserted block: %v", item)
		return true
	})

	// apply the change to the store
	err = store.ApplyChange(&space.ID, changes)
	logrus.Info(store.spaces[space.ID].blocks)
	logrus.Infof("new block: %v", store.spaces[space.ID].blocks[b1])
	//blocks, err := store.GetBlocks(&space.ID, []BlockID{b1})
	//assert.Len(t, blocks, 1)
	//assert.Equal(t, b1, blocks[0].ID)
	//assert.NoError(t, err)
}
