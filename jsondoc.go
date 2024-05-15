package blocktree

import (
	"bytes"
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/google/uuid"
	"github.com/wI2L/jsondiff"
)

type JsonDocID = uuid.UUID

type JsonPatch = []byte

// JsonDoc is a json document with incremental updates.
type JsonDoc struct {
	Content []byte `json:"string"`
}

type JsonDocPatch struct {
	ID uuid.UUID `json:"id"`
}

// DefaultJsonDoc creates a new JsonDoc.
// json docs lives in separate table
// the content structure is kept in blocks table
func DefaultJsonDoc() *JsonDoc {
	return &JsonDoc{
		Content: []byte(`{}`),
	}
}

func NewJsonDoc(content []byte) *JsonDoc {
	return &JsonDoc{
		Content: content,
	}
}

func (j *JsonDoc) Apply(patch JsonPatch) error {
	oldContent := j.Content
	if oldContent == nil {
		oldContent = []byte(`{}`)
	}
	ready, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return err
	}

	content, err := ready.Apply(oldContent)
	if err != nil {
		return err
	}

	j.Content = content

	return nil
}

// Diff computes the difference between two json documents.
func (j *JsonDoc) Diff(other *JsonDoc) (JsonPatch, error) {
	patch, err := jsondiff.CompareJSON(
		[]byte(j.Content),
		[]byte(other.Content),
		jsondiff.Factorize(),
	)
	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	return marshal, nil
}

// Clone creates a copy of the json document.
func (j *JsonDoc) Clone() *JsonDoc {
	if j == nil {
		return nil
	}

	return &JsonDoc{
		Content: bytes.Clone(j.Content),
	}
}

// String returns the string representation of the json document.
func (j *JsonDoc) String() string {
	if j == nil {
		return ""
	}

	return string(j.Content)
}

// Bytes returns the byte representation of the json document.
func (j *JsonDoc) Bytes() []byte {
	return j.Content
}
