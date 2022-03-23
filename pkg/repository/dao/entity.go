package dao

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/store"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/tdtl"
)

type Entity struct {
	ID            string               `json:"id" msgpack:"id" mapstructure:"id"`
	Type          string               `json:"type" msgpack:"type" mapstructure:"type"`
	Owner         string               `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source        string               `json:"source" msgpack:"source" mapstructure:"source"`
	Version       int64                `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime      int64                `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	TemplateID    string               `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Properties    map[string]tdtl.Node `json:"-" msgpack:"-" mapstructure:"-"`
	ConfigBytes   []byte               `json:"-" msgpack:"config_bytes" mapstructure:"config_bytes"`
	PropertyBytes []byte               `json:"property_bytes" msgpack:"property_bytes" mapstructure:"property_bytes"`
}

func Encode(en *Entity) ([]byte, error) {
	properties, err := xjson.EncodeJSON(en.Properties)
	if nil != err {
		return nil, errors.Wrap(err, "encode json")
	}

	bytes, err := json.Marshal(en)
	if nil != err {
		return nil, errors.Wrap(err, "encode json")
	}

	cc := tdtl.New(bytes)
	cc.Del("property_bytes")
	cc.Set("properties", tdtl.New(properties))
	return cc.Raw(), cc.Error()
}

func (e *Entity) Copy() Entity {
	en := Entity{
		ID:            e.ID,
		Type:          e.Type,
		Owner:         e.Owner,
		Source:        e.Source,
		Version:       e.Version,
		LastTime:      e.LastTime,
		TemplateID:    e.TemplateID,
		ConfigBytes:   []byte(`{}`),
		Properties:    make(map[string]tdtl.Node),
		PropertyBytes: []byte(`{}`),
	}

	// copy entity properties.
	for pid, pval := range e.Properties {
		en.Properties[pid] = pval
	}

	if len(en.ConfigBytes) > 0 {
		en.ConfigBytes = make([]byte, len(e.ConfigBytes))
		copy(en.ConfigBytes, e.ConfigBytes)
	}

	return en
}

func (e *Entity) Basic() Entity {
	en := Entity{
		ID:          e.ID,
		Type:        e.Type,
		Owner:       e.Owner,
		Source:      e.Source,
		Version:     e.Version,
		LastTime:    e.LastTime,
		TemplateID:  e.TemplateID,
		ConfigBytes: []byte(`{}`),
		Properties:  make(map[string]tdtl.Node),
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
	info["config_file"] = string(e.ConfigBytes)

	bytes, _ := json.Marshal(info)
	return string(bytes)
}

// PutEntity upsert Entity.
func (d *Dao) PutEntity(ctx context.Context, eid string, data []byte) error {
	err := d.stateClient.Set(ctx, d.entityCodec.Key(eid), data)
	return errors.Wrap(err, "repo put entity")
}

// GetEntity returns Entity.
func (d *Dao) GetEntity(ctx context.Context, id string) (_ []byte, err error) {
	var item *store.StateItem
	item, err = d.stateClient.Get(ctx, d.entityCodec.Key(id))
	if nil == err {
		if len(item.Value) == 0 {
			return nil, xerrors.ErrEntityNotFound
		}
		return item.Value, nil
	}
	return nil, errors.Wrap(err, "repo get entity")
}

// DelEntity delete Entity by entity id.
func (d *Dao) DelEntity(ctx context.Context, id string) error {
	return errors.Wrap(d.stateClient.Del(ctx, d.entityCodec.Key(id)), "repo del entity")
}

// HasEntity return true if entity exists, otherwise return false.
func (d *Dao) HasEntity(ctx context.Context, id string) (bool, error) {
	_, err := d.stateClient.Get(ctx, d.entityCodec.Key(id))
	if nil != err {
		if errors.Is(err, xerrors.ErrEntityNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "repo exists entity")
	}
	return true, nil
}
