/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runtime2

import (
	"context"
	"fmt"
	"github.com/tkeel-io/tdtl"
)


type EntityEventType string

const (
	OpEntityGet           EntityEventType    = "core.operation.Entity.Get"
	OpEntityPropsGet      EntityEventType    = "core.operation.Entity.Props.Get"
	OpEntityPropsUpdata   EntityEventType    = "core.operation.Entity.Props.Update"
	OpEntityPropsPatch    EntityEventType    = "core.operation.Entity.Props.Patch"
	OpEntityConfigsGet    EntityEventType    = "core.operation.Entity.Configs.Get"
	OpEntityConfigsUpdata EntityEventType    = "core.operation.Entity.Configs.Update"
	OpEntityConfigsPatch  EntityEventType    = "core.operation.Entity.Configs.Patch"
)
type EntityEvent struct {
	JSONPath string
	OP       EntityEventType
	Value    []byte
}
type mockEntity struct {
	ID         string `json:"id" msgpack:"id" mapstructure:"id"`
	Type       string `json:"type" msgpack:"type" mapstructure:"type"`
	Owner      string `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source     string `json:"source" msgpack:"source" mapstructure:"source"`
	Version    int64  `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime   int64  `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	TemplateID string `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Property   *tdtl.Collect
	Scheme     *tdtl.Collect
}

func NewEntity() Entity {
	Property := tdtl.New(`{}`)
	Scheme := tdtl.New(`{}`)
	return &mockEntity{
		Property: Property,
		Scheme:   Scheme,
	}
}

func (m *mockEntity) Handle(ctx context.Context, msg interface{}) (*StateResult, error) {
	ev, ok := msg.(*EntityEvent)
	if !ok {
		return nil, fmt.Errorf("Handle unknown type.")
	}
	switch ev.OP {
	case OpEntityGet:
		m.Property.Set(ev.JSONPath, tdtl.New(ev.Value))
		ret, err := m.Raw()
		if err != nil {
			return nil, err
		}
		return &StateResult{State: ret}, nil
	case OpEntityPropsUpdata:
		m.Property.Set(ev.JSONPath, tdtl.New(ev.Value))
		ret, err := m.Raw()
		if err != nil {
			return nil, err
		}
		return &StateResult{State: ret}, nil
	//case APIPatchEntityProps:
	//	m.Property.Set(ev.JSONPath, tdtl.New(ev.Value))
	//	ret, err := m.Raw()
	//	if err != nil {
	//		return nil, err
	//	}
	//	return &StateResult{State: ret}, nil
	case OpEntityPropsGet:
		ret, err := m.Raw()
		if err != nil {
			return nil, err
		}
		return &StateResult{State: ret}, nil
	case OpEntityConfigsUpdata:
		m.Scheme.Set(ev.JSONPath, tdtl.New(ev.Value))
		ret, err := m.Raw()
		if err != nil {
			return nil, err
		}
		return &StateResult{State: ret}, nil
	case OpEntityConfigsGet:
		ret, err := m.Raw()
		if err != nil {
			return nil, err
		}
		return &StateResult{State: ret}, nil
	}
	return nil, fmt.Errorf("Unknown EntityEvent type.")
}

func (m *mockEntity) Raw() ([]byte, error) {
	ret := tdtl.New("{}")

	ret.Set("ID", tdtl.NewString(m.ID))
	ret.Set("Type", tdtl.NewString(m.Type))
	ret.Set("Owner", tdtl.NewString(m.Owner))
	ret.Set("Source", tdtl.NewString(m.Source))
	ret.Set("Version", tdtl.NewInt64(m.Version))
	ret.Set("LastTime", tdtl.NewInt64(m.LastTime))
	ret.Set("TemplateID", tdtl.NewString(m.TemplateID))
	ret.Set("Property", m.Property)
	ret.Set("Scheme", m.Scheme)
	return ret.Raw(), m.Property.Error()
}
