package repository

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/tdtl"
)

const (
	EntityTypeBasic        = "BASIC"
	EntityTypeSubscription = "SUBSCRIPTION"
	EntityStorePrefix      = "CORE.ENTITY"
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

// --------------------------------- Storage.

type defaultEntityCodec struct {
}

func (ec *defaultEntityCodec) Key() dao.Codec {
	return &entityKeyCodec{}
}

func (ec *defaultEntityCodec) Value() dao.Codec {
	return &entityValueCodec{}
}

type entityKeyCodec struct{}

func (ec *entityKeyCodec) Encode(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case *entityResource:
		return []byte(EntityStorePrefix + "." + val.id), nil
	default:
		return nil, xerrors.ErrInternal
	}
}

func (ec *entityKeyCodec) Decode(raw []byte, v interface{}) error {
	// nerver use.
	return nil
}

type entityValueCodec struct{}

func (ec *entityValueCodec) Encode(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case *entityResource:
		return val.data, nil
	default:
		return nil, xerrors.ErrInternal
	}
}

func (ec *entityValueCodec) Decode(raw []byte, v interface{}) error {
	switch val := v.(type) {
	case *entityResource:
		val.data = raw
		return nil
	default:
		return xerrors.ErrInternal
	}
}

type entityResource struct {
	id   string
	data []byte
}

func (e *entityResource) Codec() dao.KVCodec {
	return &defaultEntityCodec{}
}

func (r *repo) PutEntity(ctx context.Context, eid string, data []byte) error {
	err := r.dao.StoreResource(ctx, &entityResource{id: eid, data: data})
	return errors.Wrap(err, "put entity repository")
}

func (r *repo) GetEntity(ctx context.Context, eid string) ([]byte, error) {
	ret, err := r.dao.GetStoreResource(ctx, &entityResource{id: eid})

	res, _ := ret.(*entityResource)
	return res.data, errors.Wrap(err, "get entity repository")
}

func (r *repo) DelEntity(ctx context.Context, eid string) error {
	err := r.dao.RemoveStoreResource(ctx, &entityResource{id: eid})
	return errors.Wrap(err, "del entity repository")
}

func (r *repo) HasEntity(ctx context.Context, eid string) (bool, error) {
	_, err := r.dao.GetStoreResource(ctx, &entityResource{id: eid})
	if nil != err {
		if errors.Is(err, xerrors.ErrResourceNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "exists entity repository")
	}
	return true, errors.Wrap(err, "exists entity repository")
}
