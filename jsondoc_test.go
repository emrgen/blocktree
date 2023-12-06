package blocktree

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	d1, _ = uuid.Parse("00000000-0000-0000-0000-000000000001")
	d2, _ = uuid.Parse("00000000-0000-0000-0000-000000000002")
	d3, _ = uuid.Parse("00000000-0000-0000-0000-000000000003")
)

func TestNewJsonDoc(t *testing.T) {
	jd1 := NewJsonDoc(d1)
	jd1.Content = `{"name":"John Doe"}`

	jd2 := NewJsonDoc(d2)
	jd2.Content = `{"name":"tommy","age":12}`

	diff, err := jd1.Diff(jd2)
	assert.NoError(t, err)
	if err != nil {
		logrus.Errorf("decode patch error: %v", err)
	}
	logrus.Infof("diff: %v", string(diff))

	err = jd1.ApplyPatch(diff)
	assert.NoError(t, err)

	assert.Equal(t, jd1.Content, jd2.Content)
}
