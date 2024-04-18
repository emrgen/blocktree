package blocktree

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMoveTree_Move(t *testing.T) {
	s := uuid.New()
	tree := newMoveTree(s)
	err := tree.move(s, s)
	assert.Error(t, err)

	a := uuid.New()
	b := uuid.New()
	tree.addEdge(a, s)
	tree.addEdge(b, s)
	//tree.print()

	err = tree.move(a, b)
	assert.NoError(t, err)
	//tree.print()

	err = tree.move(b, a)
	assert.Equal(t, err, ErrCreatesCycle)

	err = tree.move(a, s)
	assert.NoError(t, err)

	c := uuid.New()
	tree.addEdge(c, a)

	d := uuid.New()
	tree.addEdge(d, c)
}
