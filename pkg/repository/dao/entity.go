package dao

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/store"
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
	ConfigFile []byte                     `json:"-" msgpack:"config_file" mapstructure:"config_file"`
}

func (e *Entity) Copy() Entity {
	en := Entity{
		ID:         e.ID,
		Type:       e.Type,
		Owner:      e.Owner,
		Source:     e.Source,
		Version:    e.Version,
		LastTime:   e.LastTime,
		TemplateID: e.TemplateID,
		Properties: make(map[string]constraint.Node),
	}

	// copy entity properties.
	for pid, pval := range e.Properties {
		en.Properties[pid] = pval.Copy()
	}

	copy(en.ConfigFile, e.ConfigFile)
	return en
}

func (e *Entity) Basic() Entity {
	en := Entity{
		ID:         e.ID,
		Type:       e.Type,
		Owner:      e.Owner,
		Source:     e.Source,
		Version:    e.Version,
		LastTime:   e.LastTime,
		TemplateID: e.TemplateID,
		Properties: make(map[string]constraint.Node),
	}

	return en
}

func (e *Entity) JSON() string {
	info := make(map[string]interface{})
	info["id"] = e.ID
	info["type"] = e.Type
	info["owner"] = e.Owner
	info["source"] = e.Source
	info["version"] = e.Version
	info["last_time"] = e.LastTime
	info["template_id"] = e.TemplateID

	// marhsal properties.
	props := make(map[string]interface{})
	for key, val := range e.Properties {
		props[key] = val.String()
	}

	info["properties"] = props
	info["config_file"] = string(e.ConfigFile)

	bytes, _ := json.Marshal(info)
	return string(bytes)
}

// dao interfaces.
func (d *Dao) PutEntity(ctx context.Context, en *Entity) error {
	var err error
	var bytes []byte
	if bytes, err = d.entityCodec.Encode(en); nil == err {
		err = d.stateClient.Set(ctx, d.entityCodec.Key(en.ID), bytes)
	}
	return errors.Wrap(err, "repo put entity")
}

func (d *Dao) GetEntity(ctx context.Context, id string) (en *Entity, err error) {
	var item *store.StateItem
	item, err = d.stateClient.Get(ctx, d.entityCodec.Key(id))
	if nil == err {
		if len(item.Value) == 0 {
			return nil, xerrors.ErrEntityNotFound
		}

		en = new(Entity)
		err = d.entityCodec.Decode(item.Value, en)
	}
	return en, errors.Wrap(err, "repo get entity")
}

func (d *Dao) DelEntity(ctx context.Context, id string) error {
	return errors.Wrap(d.stateClient.Del(ctx, d.entityCodec.Key(id)), "repo del entity")
}

func (d *Dao) HasEntity(ctx context.Context, id string) (bool, error) {
	res, err := d.stateClient.Get(ctx, d.entityCodec.Key(id))
	if nil == err {
		if len(res.Value) > 0 {
			return true, nil
		}
	}

	return false, errors.Wrap(err, "repo exists entity")
}
