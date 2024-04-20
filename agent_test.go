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

	// create a block agent
	agents := []*blockAgent{
		newBlockAgent(aid1, sid, NewMemStore(), server),
		newBlockAgent(aid2, sid, NewMemStore(), server),
	}

	simulateAgents(agents)

	// check if all agents have the same block tree
	for i := 0; i < len(agents); i++ {
		for j := 0; j < len(agents); j++ {
			assert.Equal(t, agents[i].equalState(agents[j]), true)
		}
	}
}
