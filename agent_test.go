package blocktree

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func TestAgentsInsert(t *testing.T) {
	// create a block server
	server := newBlockServer()

	aid1 := uuid.New()
	aid2 := uuid.New()
	sid := uuid.New()

	// create two agents
	agents := []*blockAgent{
		newBlockAgent(aid1, sid, NewMemStore(), server),
		newBlockAgent(aid2, sid, NewMemStore(), server),
	}

	// start agents and simulate insert operations
	simulateAgents(agents, NewSet[OpType](OpTypeInsert))

	// check if all agents have the same block tree
	for i := 0; i < len(agents); i++ {
		for j := 0; j < len(agents); j++ {
			assert.Equal(t, agents[i].equalState(agents[j]), true)
		}
	}
}

func TestSyncAgents1(t *testing.T) {
	// create a block server
	server := newBlockServer()

	aid1 := uuid.New()
	aid2 := uuid.New()

	a1 := newBlockAgent(aid1, s1, NewMemStore(), server)
	a2 := newBlockAgent(aid2, s1, NewMemStore(), server)

	a1tx1 := createTx(s1, insertOp(b1, "p1", s1, PositionStart))

	_, err := a1.api.Apply(a1tx1)
	assert.NoError(t, err)

	block, err := a1.api.GetBlock(s1, b1)
	assert.NoError(t, err)

	assert.Equal(t, block.ID, b1)

	a1.api.store.(*MemStore).Print(&s1)

	agents := []*blockAgent{a1, a2}
	// check if all agents have the same block tree
	for i := 0; i < len(agents); i++ {
		for j := 0; j < len(agents); j++ {
			assert.Equal(t, agents[i].equalState(agents[j]), true)
		}
	}
}
