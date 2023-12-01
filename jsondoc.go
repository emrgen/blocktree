package blocktree

import "github.com/google/uuid"

type JsonDocID = uuid.UUID

// JsonDoc is a json document with incremental updates.
type JsonDoc struct {
	ID      JsonDocID              `json:"id"`
	Content map[string]interface{} `json:"content"`
}

type JsonDocPatch struct {
	ID  uuid.UUID        `json:"id"`
	Ops []JsonDocPatchOp `json:"ops"`
}

type JsonDocPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// NewJsonDoc creates a new JsonDoc.
func NewJsonDoc(id uuid.UUID) *JsonDoc {
	return &JsonDoc{
		ID:      id,
		Content: make(map[string]interface{}),
	}
}

func (j *JsonDoc) Patch(patch JsonDocPatch) error {
	return nil
}

// JsonDocChange is a change to a json document.
type JsonDocChange struct {
	change []*JsonDoc
}
