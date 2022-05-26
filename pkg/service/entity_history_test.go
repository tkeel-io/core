package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/store"
	_ "github.com/tkeel-io/core/pkg/resource/store/memory"
)

func Test_entityHistory_AddEnity(t *testing.T) {
	type fields struct {
		count int
		store store.Store
	}
	type args struct {
		user     string
		entityID string
	}
	metadata := resource.Metadata{Name: "memory"}
	storeTest := store.NewStore(metadata)
	tests := []struct {
		name   string
		fields fields
		args   args
		wants  []string
	}{
		// TODO: Add test cases.
		{"1", fields{5, storeTest}, args{"u1", "d1"}, []string{"d1"}},
		{"2", fields{5, storeTest}, args{"u1", "d1"}, []string{"d1"}},
		{"3", fields{5, storeTest}, args{"u1", "d2"}, []string{"d2", "d1"}},
		{"4", fields{5, storeTest}, args{"u1", "d3"}, []string{"d3", "d2", "d1"}},
		{"5", fields{5, storeTest}, args{"u1", "d4"}, []string{"d4", "d3", "d2", "d1"}},
		{"6", fields{5, storeTest}, args{"u1", "d5"}, []string{"d5", "d4", "d3", "d2", "d1"}},
		{"7", fields{5, storeTest}, args{"u1", "d6"}, []string{"d6", "d5", "d4", "d3", "d2"}},
		{"8", fields{5, storeTest}, args{"u1", "d2"}, []string{"d2", "d6", "d5", "d4", "d3"}},
		{"9", fields{5, storeTest}, args{"u1", "d2"}, []string{"d2", "d6", "d5", "d4", "d3"}},
		{"10", fields{5, storeTest}, args{"u2", "d1"}, []string{"d1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &entityHistory{
				count: tt.fields.count,
				store: tt.fields.store,
			}
			e.AddEnity(tt.args.user, tt.args.entityID)
			entityList := e.GetLatestEntities(tt.args.user)
			t.Log(entityList)
			assert.Equal(t, tt.wants, entityList)
		})
	}
	e := &entityHistory{
		count: 5,
		store: storeTest,
	}
	entityList := e.GetLatestEntities("u3")
	for range entityList {
	}
	t.Log(entityList)
	var want []string
	assert.Equal(t, want, entityList)
	t.Log()
}
