package dao

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/resource/state"
)

type Entity struct {
	ID         string                     `json:"id" msgpack:"id" mapstructure:"id"`
	Type       string                     `json:"type" msgpack:"type" mapstructure:"type"`
	Owner      string                     `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source     string                     `json:"source" msgpack:"source" mapstructure:"source"`
	Version    int64                      `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime   int64                      `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	TemplateID string                     `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Properties map[string]constraint.Node `json:"properties" msgpack:"properties" mapstructure:"-"`
	ConfigFile []byte                     `json:"-" msgpack:"config_file" mapstructure:"-"`
}

func (d *Dao) Put(ctx context.Context, en *Entity) error {
	bytes, err := Encode(en)
	if nil == err {
		err = d.stateClient.Set(ctx, StoreKey(en.ID), bytes)
	}
	return errors.Wrap(err, "put entity")
}

func (d *Dao) Get(ctx context.Context, id string) (en *Entity, err error) {
	var item *state.StateItem
	item, err = d.stateClient.Get(ctx, StoreKey(id))
	if nil == err {
		en = new(Entity)
		err = Decode(item.Value, en)
	}
	return en, errors.Wrap(err, "get entity")
}
