package constraint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	valInt := 23
	valFloat := 55.8

	assert.Equal(t, NewNode(true), BoolNode(true), "BoolNode<true>.")
	assert.Equal(t, NewNode(false), BoolNode(false), "BoolNode<false>.")
	assert.Equal(t, NewNode(int(1)), IntNode(1), "IntNode.")
	assert.Equal(t, NewNode(1.2), FloatNode(1.2), "FloatNode.")
	assert.Equal(t, NewNode(-22.1).To(String), StringNode("-22.1"), "StringNode.")
	assert.Equal(t, NewNode(&valInt), IntNode(valInt), "IntNode PTR.")
	assert.Equal(t, NewNode(&valFloat), FloatNode(valFloat), "FloatNode PTR.")
	assert.Equal(t, NewNode([]byte("test bytes.")), JSONNode("test bytes."), "RawNode PTR.")

	t.Log(NewNode(-22.1).To(String).String())

	time.Sleep(time.Second)
}
