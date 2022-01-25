package tql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTQL(t *testing.T) {
	tqlInst, err := NewTQL("insert into iotd-1773_core-broker-0 select iotd-1773_.*")

	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"iotd-1773_"}, tqlInst.Entities())
	assert.Equal(t, "iotd-1773_core-broker-0", tqlInst.Target())
	assert.Equal(t, []TentacleConfig{{SourceEntity: "iotd-1773_", PropertyKeys: []string{"*"}}}, tqlInst.Tentacles())
}
