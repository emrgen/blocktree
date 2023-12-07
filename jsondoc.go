package blocktree

import (
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/google/uuid"
)

type JsonDocID = uuid.UUID

type JsonPatch = []byte

// JsonDoc is a json document with incremental updates.
type JsonDoc struct {
	Content []byte `json:"string"`
}

type JsonDocPatch struct {
	ID uuid.UUID `json:"id"`
	//Ops []JsonDocPatchOp `json:"ops"`
}

// NewJsonDoc creates a new JsonDoc.
// json docs lives in separate table
// the content structure is kept in blocks table
func NewJsonDoc() *JsonDoc {
	return &JsonDoc{
		Content: []byte(`{}`),
	}
}

func (j *JsonDoc) Apply(patch JsonPatch) error {
	content, err := jsonpatch.MergePatch(j.Content, patch)
	if err != nil {
		return err
	}
	j.Content = content

	return nil
}

func (j *JsonDoc) Diff(other *JsonDoc) (JsonPatch, error) {
	return jsonpatch.CreateMergePatch([]byte(j.Content), []byte(other.Content))
}
