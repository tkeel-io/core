package dao

import (
	"github.com/tkeel-io/core/pkg/constraint"
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
