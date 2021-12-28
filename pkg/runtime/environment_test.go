package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironment(t *testing.T) {
	env := NewEnv()

	infos := env.LoadMapper([]EtcdPair{
		{
			Key:   "core.mappe.BASICr.device123.mapper-from-device234",
			Value: []byte("insert into device123 select device234.temp as temp"),
		},
		{
			Key:   "core.mapper.SUBSCRIPTION.sub123.mapper-from-device234",
			Value: []byte("insert into sub123 select device234.temp"),
		},
	})

	assert.Equal(t, "sub123", infos[0].EntityID)
}
