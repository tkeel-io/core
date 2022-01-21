package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func TestEnvironment(t *testing.T) {
	env := NewEnvironment()

	infos := env.StoreMappers([]dao.Mapper{
		{
			ID:       "mapper123",
			Name:     "mapping device234 props",
			EntityID: "device123",
			TQL:      "insert into device123 select device234.temp as temp",
		},
		{
			ID:       "mapper234",
			Name:     "mapping device234 props",
			EntityID: "sub123",
			TQL:      "insert into sub123 select device234.temp",
		},
	})

	assert.Equal(t, "device123", infos[0].EntityID)
}
