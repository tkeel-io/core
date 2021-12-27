package tql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTQL(t *testing.T) {
	tqlInst, err := NewTQL("insert into iotd-1773_core-broker-0 select iotd-1773_.*")
	assert.Equal(t, nil, err)
	t.Log(tqlInst.Target())
}
