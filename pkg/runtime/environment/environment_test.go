package environment

import (
	"sort"
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

	stateEnv := env.GetStateEnv("device123")
	t.Log("mappers", stateEnv.Mappers)
	for _, tentacle := range stateEnv.Tentacles {
		t.Log("tentacle", tentacle)
	}

	stateEnv = env.GetStateEnv("device234")
	t.Log("mappers", stateEnv.Mappers)
	for _, tentacle := range stateEnv.Tentacles {
		t.Log("tentacle", tentacle)
	}

	effect, _ := env.OnMapperChanged(dao.PUT, dao.Mapper{
		ID:       "mapper123",
		Name:     "mapping device234 props",
		EntityID: "device123",
		TQL:      "insert into device123 select device234.temp as temp",
	})

	t.Log("effect: ", effect)

	effect, _ = env.OnMapperChanged(dao.PUT, dao.Mapper{
		ID:       "mapper234",
		Name:     "mapping device234 props",
		EntityID: "sub123",
		TQL:      "insert into sub123 select device234.temp",
	})

	t.Log("effect: ", effect)

	assert.Equal(t, "device123", infos[0].EntityID)
}

func TestSlice(t *testing.T) {
	slice := sort.StringSlice{"device234", "device123"}
	slice.Sort()
	t.Log(slice)
	t.Log(slice.Search("device3"))
}
