package blocktree

import (
	"reflect"
	"sync"

	"github.com/google/uuid"
)

type blockServer struct {
	api *Api
	mu  sync.Mutex
}

func newBlockServer() *blockServer {
	store := NewMemStore()
	return &blockServer{
		api: NewApi(store),
	}
}

type blockAgent struct {
	id       uuid.UUID
	spaceID  uuid.UUID
	server   *blockServer
	api      *Api
	blocks   map[string]*Block
	children map[string][]string
}

func newBlockAgent(id uuid.UUID, spaceID SpaceID, store Store, server *blockServer) *blockAgent {
	api := NewApi(NewMemStore())
	err := api.CreateSpace(spaceID, spaceID.String())
	if err != nil {
		panic(err)
	}

	return &blockAgent{
		id:       id,
		spaceID:  spaceID,
		api:      api,
		server:   server,
		blocks:   make(map[string]*Block),
		children: make(map[string][]string),
	}
}

func (a *blockAgent) start() {
}

func (a *blockAgent) stop() {

}

func (a *blockAgent) equalState(other *blockAgent) bool {
	if a.spaceID != other.spaceID {
		return false
	}

	// compare blocks
	if len(a.blocks) != len(other.blocks) {
		return false
	}

	for id, block := range a.blocks {
		otherBlock, ok := other.blocks[id]
		if !ok || !reflect.DeepEqual(block, otherBlock) {
			return false
		}
	}

	// compare children
	if len(a.children) != len(other.children) {
		return false
	}

	for id, children := range a.children {
		otherChildren, ok := other.children[id]
		if !ok || !reflect.DeepEqual(children, otherChildren) {
			return false
		}
	}

	return true
}

// agents updates the local state separately and tries to reach consensus
// by comparing the local state with the state of other agents.
func simulateAgents(agents []*blockAgent, ops *Set[OpType]) {
	//for _, agent := range agents {
	//	agent.start()
	//}
	//
	//for _, agent := range agents {
	//	agent.stop()
	//}
}
