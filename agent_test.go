package blocktree

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func TestAgentsInsert(t *testing.T) {
	// create a block server
	server := newBlockServer()

	aid1 := uuid.New().String()
	aid2 := uuid.New().String()
	sid := uuid.New().String()

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

	aid1 := uuid.New().String()
	aid2 := uuid.New().String()
	sid := uuid.New().String()

	a1 := newBlockAgent(aid1, sid, NewMemStore(), server)
	a2 := newBlockAgent(aid2, sid, NewMemStore(), server)

	agents := []*blockAgent{a1, a2}
	// check if all agents have the same block tree
	for i := 0; i < len(agents); i++ {
		for j := 0; j < len(agents); j++ {
			assert.Equal(t, agents[i].equalState(agents[j]), true)
		}
	}
}
