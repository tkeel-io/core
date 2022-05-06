package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MemStore(t *testing.T) {
	val := []byte("{}")
	ns, err := initStore(nil)
	assert.Nil(t, err)
	ret, err := ns.Get(context.Background(), "entity123")
	assert.Nil(t, ret)
	assert.NotNil(t, err)
	err = ns.Set(context.Background(), "entity123", []byte("{}"))
	assert.Nil(t, err)
	ret, err = ns.Get(context.Background(), "entity123")
	assert.Equal(t, ret.Value, val)
	assert.Nil(t, err)
	err = ns.Del(context.Background(), "entity123")
	assert.Nil(t, err)
	ret, err = ns.Get(context.Background(), "entity123")
	assert.NotNil(t, err)
	assert.Nil(t, ret)
}
