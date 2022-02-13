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

package state

// statem: state machine.

import (
	"context"
	"sort"

	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/dispatch"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
)

const (
	// statem runtime-status enumerates.
	StateRuntimeDetached int32 = 0
	StateRuntimeAttached int32 = 1

	// statem status enumerates.
	SMStatusActive   Status = "active"
	SMStatusInactive Status = "inactive"
	SMStatusDeleted  Status = "deleted"

	// reserved property field.
	ReservedFieldID         = "id"
	ReservedFieldType       = "type"
	ReservedFieldOwner      = "owner"
	ReservedFieldSource     = "source"
	ReservedFieldVersion    = "version"
	ReservedFieldConfigs    = "configs"
	ReservedFieldMappers    = "mappers"
	ReservedFieldLastTime   = "last_time"
	ReservedFieldTemplate   = "template"
	ReservedFieldProperties = "properties"
	ReservedFieldConfigFile = "config_file"
)

var RequiredFields = map[string]bool{
	ReservedFieldID:         true,
	ReservedFieldType:       true,
	ReservedFieldOwner:      true,
	ReservedFieldSource:     true,
	ReservedFieldVersion:    false,
	ReservedFieldConfigs:    false,
	ReservedFieldMappers:    false,
	ReservedFieldLastTime:   false,
	ReservedFieldTemplate:   false,
	ReservedFieldProperties: false,
	ReservedFieldConfigFile: false,
}

var squashFields = map[string]string{
	ReservedFieldID:     message.ExtEntityID,
	ReservedFieldType:   message.ExtEntityType,
	ReservedFieldOwner:  message.ExtEntityOwner,
	ReservedFieldSource: message.ExtEntitySource,
}

func SquashFields(header map[string]string) map[string]string {
	ret := make(map[string]string)
	for key, value := range header {
		ret[squashFields[key]] = value
	}
	return ret
}

// statem state marchins.
type statem struct {
	// state basic fields.
	dao.Entity
	// other state machine property cache.
	cacheProps map[string]map[string]constraint.Node // cache other property.

	// mapper & tentacles.
	mappers   map[string]mapper.Mapper      // key=mapperId
	tentacles map[string][]mapper.Tentacler // key=Sid#propertyKey

	// parse from Configs.
	constraints        map[string]*constraint.Constraint
	searchConstraints  sort.StringSlice
	tseriesConstraints sort.StringSlice

	// state manager.
	dispatcher      dispatch.Dispatcher
	resourceManager types.ResourceManager

	status Status
	// state machine message handler.
	msgHandler MessageHandler

	// state Context.
	sCtx   StateContext
	ctx    context.Context
	cancel context.CancelFunc
}

// NewState create an statem object.
func NewState(ctx context.Context, in *dao.Entity, dispatcher dispatch.Dispatcher, resourceManager types.ResourceManager, msgHandler MessageHandler) (Machiner, error) {
	if in.ID == "" {
		in.ID = util.UUID()
	}

	ctx, cancel := context.WithCancel(ctx)

	state := &statem{
		Entity: in.Copy(),

		ctx:             ctx,
		cancel:          cancel,
		status:          SMStatusActive,
		msgHandler:      msgHandler,
		dispatcher:      dispatcher,
		resourceManager: resourceManager,
		mappers:         make(map[string]mapper.Mapper),
		cacheProps:      make(map[string]map[string]constraint.Node),
		constraints:     make(map[string]*constraint.Constraint),
	}

	// initialize Properties.
	if nil == state.Entity.Properties {
		state.Properties = make(map[string]constraint.Node)
	}

	// set properties into cacheProps.
	state.cacheProps[in.ID] = state.Properties

	state.msgHandler = state.invokePropertyMessage

	return state, nil
}

// GetID returns state ID.
func (s *statem) GetID() string {
	return s.ID
}

// GetStatus returns state machine status.
func (s *statem) GetStatus() Status {
	return s.status
}

func (s *statem) GetEntity() *dao.Entity {
	return &s.Entity
}

// WithContext set state Context.
func (s *statem) WithContext(ctx StateContext) Machiner {
	s.sCtx = ctx
	return s
}

// OnMessage recive statem input messages.
func (s *statem) Invoke(msgCtx message.Context) error {
	switch msgCtx.Message().(type) {
	case message.StateMessage:
		actives := s.callAPIs(context.Background(), msgCtx)
		s.activeTentacle(actives)
	default:
		// handle message.
		watchKeys := s.msgHandler(msgCtx)
		// active tentacles.
		s.activeTentacle(watchKeys)
	}
	return nil
}

type Mapper struct {
	ID          string `json:"id" msgpack:"id"`
	TQL         string `json:"tql" msgpack:"tql"`
	Name        string `json:"name" msgpack:"name"`
	Description string `json:"description" msgpack:"description"`
}
