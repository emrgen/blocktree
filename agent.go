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

// newBlockServer creates a new block server with a space.
func newBlockServer(spaceID SpaceID) *blockServer {
	store := NewMemStore()
	api := NewApi(store)
	api.CreateSpace(spaceID, spaceID.String())
	return &blockServer{
		api: api,
	}
}

// apply applies transactions to the server.
func (s *blockServer) apply(tx ...*Transaction) (*SyncBlocks, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.api.Apply(tx...)
}

type blockAgent struct {
	id       uuid.UUID
	spaceID  uuid.UUID
	server   *blockServer
	api      *Api
	blocks   map[string]*Block
	children map[string][]string
	applied  []*Transaction
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
		applied:  make([]*Transaction, 0),
	}
}

func (a *blockAgent) apply(tx ...*Transaction) error {
	_, err := a.api.Apply(tx...)
	if err != nil {
		return err
	}

	a.applied = append(a.applied, tx...)

	return nil
}

// sync applies all transactions that were not applied yet.
func (a *blockAgent) sync(server *blockServer) error {
	applied := a.applied
	for _, tx := range a.applied {
		_, err := server.apply(tx)
		if err != nil {
			return err
		}

		applied = applied[1:]
	}

	a.applied = applied

	return nil
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
