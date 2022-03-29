package manager

import (
	"encoding/json"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/tdtl"
)

// EntityBase state basic informatinon.
type Base struct {
	ID         string       `json:"id" msgpack:"id" mapstructure:"id"`
	Type       string       `json:"type" msgpack:"type" mapstructure:"type"`
	Owner      string       `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source     string       `json:"source" msgpack:"source" mapstructure:"source"`
	Version    int64        `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime   int64        `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	Mappers    []*v1.Mapper `json:"mappers" msgpack:"mappers" mapstructure:"mappers"`
	TemplateID string       `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Scheme     []byte       `json:"-" msgpack:"scheme" mapstructure:"-"`
	Properties []byte       `json:"properties" msgpack:"properties" mapstructure:"properties"`
}

type BaseRet struct {
	ID         string                 `json:"id" msgpack:"id" mapstructure:"id"`
	Type       string                 `json:"type" msgpack:"type" mapstructure:"type"`
	Owner      string                 `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source     string                 `json:"source" msgpack:"source" mapstructure:"source"`
	Version    int64                  `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime   int64                  `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	Mappers    []*v1.Mapper           `json:"mappers" msgpack:"mappers" mapstructure:"mappers"`
	TemplateID string                 `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Properties map[string]interface{} `json:"properties" msgpack:"properties" mapstructure:"properties"`
	Scheme     map[string]interface{} `json:"scheme" msgpack:"-" mapstructure:"scheme"`
}

func (b *Base) Basic() Base {
	cp := Base{
		ID:         b.ID,
		Type:       b.Type,
		Owner:      b.Owner,
		Source:     b.Source,
		Version:    b.Version,
		LastTime:   b.LastTime,
		TemplateID: b.TemplateID,
		Scheme:     []byte(`{}`),
		Properties: []byte(`{}`),
	}

	cp.Mappers = append(cp.Mappers, b.Mappers...)
	return cp
}

func (b *Base) JSON() map[string]interface{} {
	info := make(map[string]interface{})
	info["id"] = b.ID
	info["type"] = b.Type
	info["owner"] = b.Owner
	info["source"] = b.Source
	info["version"] = b.Version
	info["last_time"] = b.LastTime
	info["template_id"] = b.TemplateID
	info["scheme"] = string(b.Scheme)
	info["properties"] = string(b.Properties)
	return info
}

func (b *Base) EncodeJSON() ([]byte, error) {
	bytes, err := json.Marshal(b)
	if nil != err {
		return nil, errors.Wrap(err, "json encode base")
	}

	// encode scheme.
	cc := tdtl.New(bytes)
	if len(b.Scheme) > 0 {
		cc.Set("scheme", tdtl.New(b.Scheme))
	}

	// encode properties.
	if len(b.Properties) > 0 {
		cc.Set("properties", tdtl.New(b.Properties))
	}
	return cc.Raw(), errors.Wrap(err, "encode base")
}
