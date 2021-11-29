package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlag(t *testing.T) {
	bb := NewBitBucket(555)

	bb.Enable(223)
	bb.Enable(453)
	assert.Equal(t, true, bb.Enabled(223))
	assert.Equal(t, true, bb.Enabled(453))
	assert.Equal(t, false, bb.Enabled(33))
}
