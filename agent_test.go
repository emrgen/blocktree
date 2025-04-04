package blocktree

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func TestAgentsInsert(t *testing.T) {
	// create a block server
	server, _ := newBlockServer(s1)

	aid1 := uuid.New()
	aid2 := uuid.New()

	// create two agents
	agents := []*blockAgent{
		newBlockAgent(aid1, s1, NewMemStore(), server),
		newBlockAgent(aid2, s1, NewMemStore(), server),
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

func TestSyncAgentWithServer(t *testing.T) {
	var block *Block
	var err error
	// create a block server
	server, _ := newBlockServer(s1)

	aid1 := uuid.New()

	a1 := newBlockAgent(aid1, s1, NewMemStore(), server)

	a1tx1 := createTx(s1, insertOp(b1, "p1", s1, PositionStart))

	err = a1.apply(a1tx1)
	assert.NoError(t, err)

	block, err = a1.api.GetBlock(s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, block.ID, b1)

	//a1.api.store.(*MemStore).Print(&s1)

	err = a1.sync(server)
	assert.NoError(t, err)
	assert.True(t, a1.api.store.(*MemStore).Equals(server.api.store.(*MemStore)))
}

func TestSyncAgentWithServerWithConflict(t *testing.T) {
	var block *Block
	var err error
	// create a block server
	server, _ := newBlockServer(s1)

	aid1 := uuid.New()
	a1 := newBlockAgent(aid1, s1, NewMemStore(), server)

	a1tx1 := createTx(s1, insertOp(b1, "p1", s1, PositionStart))

	err = a1.apply(a1tx1)
	assert.NoError(t, err)

	block, err = a1.api.GetBlock(s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, block.ID, b1)

	//a1.api.store.(*MemStore).Print(&s1)

	// create a new agent
	aid2 := uuid.New()
	a2 := newBlockAgent(aid2, s1, NewMemStore(), server)

	a2tx1 := createTx(s1, insertOp(b1, "p1", s1, PositionEnd))

	err = a2.apply(a2tx1)
	assert.NoError(t, err)

	block, err = a2.api.GetBlock(s1, b1)
	assert.NoError(t, err)
	assert.Equal(t, block.ID, b1)

	//a2.api.store.(*MemStore).Print(&s1)

	err = a1.sync(server)
	assert.NoError(t, err)

	assert.True(t, a1.api.store.(*MemStore).Equals(server.api.store.(*MemStore)))

	err = a2.sync(server)
	assert.NoError(t, err)

	server.sync(a2)

	assert.True(t, a2.api.store.(*MemStore).Equals(a1.api.store.(*MemStore)))
}
