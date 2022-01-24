package entities

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/runtime/state"
)

// EntityBase state basic informatinon.
type Base struct {
	ID         string                       `json:"id" msgpack:"id" mapstructure:"id"`
	Type       string                       `json:"type" msgpack:"type" mapstructure:"type"`
	Owner      string                       `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source     string                       `json:"source" msgpack:"source" mapstructure:"source"`
	Version    int64                        `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime   int64                        `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	Mappers    []state.Mapper               `json:"mappers" msgpack:"mappers" mapstructure:"mappers"`
	TemplateID string                       `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Properties map[string]constraint.Node   `json:"properties" msgpack:"properties" mapstructure:"-"`
	Configs    map[string]constraint.Config `json:"configs" msgpack:"-" mapstructure:"-"`
	ConfigFile []byte                       `json:"-" msgpack:"config_file" mapstructure:"-"`
}

func (b *Base) Basic() Base {
	cp := Base{
		ID:         b.ID,
		Type:       b.Type,
		Owner:      b.Owner,
		Source:     b.Source,
		Version:    b.Version,
		LastTime:   b.LastTime,
		Properties: make(map[string]constraint.Node),
		Configs:    make(map[string]constraint.Config),
	}

	cp.Mappers = append(cp.Mappers, b.Mappers...)
	return cp
}

func (b *Base) GetProperty(path string) (constraint.Node, error) {
	if !strings.ContainsAny(path, ".[") {
		if _, has := b.Properties[path]; !has {
			return constraint.NullNode{}, xerrors.ErrPropertyNotFound
		}
		return b.Properties[path], nil
	}

	// patch copy property.
	arr := strings.SplitN(path, ".", 2)
	res, err := constraint.Patch(b.Properties[arr[0]], nil, arr[1], constraint.PatchOpCopy)
	return res, errors.Wrap(err, "patch copy")
}

func (b *Base) GetConfig(path string) (cfg constraint.Config, err error) {
	segs := strings.Split(strings.TrimSpace(path), ".")
	if len(segs) > 1 {
		// check path.
		for _, seg := range segs {
			if strings.TrimSpace(seg) == "" {
				return cfg, constraint.ErrPatchPathInvalid
			}
		}

		rootCfg, ok := b.Configs[segs[0]]
		if !ok {
			return cfg, errors.Wrap(constraint.ErrPatchPathInvalid, "root config not found")
		}

		_, pcfg, err := rootCfg.GetConfig(segs, 1)
		return *pcfg, errors.Wrap(err, "prev config not found")
	} else if len(segs) == 1 {
		if _, ok := b.Configs[segs[0]]; !ok {
			return cfg, xerrors.ErrPropertyNotFound
		}
		return b.Configs[segs[0]], nil
	}
	return cfg, errors.Wrap(constraint.ErrPatchPathInvalid, "copy config")
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
