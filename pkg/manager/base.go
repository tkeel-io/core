package manager

import (
	"encoding/json"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/tdtl"
)

// EntityBase state basic informatinon.
type Base struct {
	ID         string                        `json:"id" msgpack:"id" mapstructure:"id"`
	Type       string                        `json:"type" msgpack:"type" mapstructure:"type"`
	Owner      string                        `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source     string                        `json:"source" msgpack:"source" mapstructure:"source"`
	Version    int64                         `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime   int64                         `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	Mappers    []*v1.Mapper                  `json:"mappers" msgpack:"mappers" mapstructure:"mappers"`
	TemplateID string                        `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Properties map[string]tdtl.Node          `json:"properties" msgpack:"properties" mapstructure:"-"`
	Configs    map[string]*constraint.Config `json:"configs" msgpack:"-" mapstructure:"-"`
	ConfigFile []byte                        `json:"-" msgpack:"config_file" mapstructure:"-"`
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
	Properties map[string]interface{} `json:"properties" msgpack:"properties" mapstructure:"-"`
	Configs    map[string]interface{} `json:"configs" msgpack:"-" mapstructure:"-"`
}

func (b *Base) Basic() Base {
	cp := Base{
		ID:         b.ID,
		Type:       b.Type,
		Owner:      b.Owner,
		Source:     b.Source,
		Version:    b.Version,
		LastTime:   b.LastTime,
		Properties: make(map[string]tdtl.Node),
		Configs:    make(map[string]*constraint.Config),
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

	// marhsal properties.
	props := make(map[string]string)
	for key, val := range b.Properties {
		props[key] = val.String()
	}

	bytes, _ := json.Marshal(b.Configs)
	info["properties"] = props
	info["configs"] = string(bytes)
	info["config_file"] = string(b.ConfigFile)
	return info
}

func entityToBase(en *dao.Entity) *Base {
	base := &Base{
		ID:         en.ID,
		Type:       en.Type,
		Owner:      en.Owner,
		Source:     en.Source,
		Version:    en.Version,
		LastTime:   en.LastTime,
		TemplateID: en.TemplateID,
		Properties: en.Properties,
		ConfigFile: en.ConfigBytes,
	}

	return base
}
