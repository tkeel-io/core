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
	"strings"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
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

	// state Context, state context version.
	sCtx    StateContext
	version int64

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

	// initialize state context.
	state.sCtx = newContext(state)

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

func (s *statem) Context() *StateContext {
	return &s.sCtx
}

// WithContext set state Context.
func (s *statem) WithContext(ctx StateContext) Machiner {
	s.sCtx = ctx
	return s
}

func (s *statem) updateFromContext() {
	if atomic.LoadInt64(&s.sCtx.version) > s.version {
		log.Debug("state context changed", zfield.Eid(s.ID), zfield.Type(s.Type))
		s.mappers = s.sCtx.mappers
		s.tentacles = s.sCtx.tentacles
	}
}

// OnMessage recive statem input messages.
func (s *statem) Invoke(msgCtx message.Context) error {
	// update state from StateContext.
	s.updateFromContext()

	// delive message.
	msgType := msgCtx.Get(message.ExtMessageType)
	switch message.MessageType(msgType) {
	case message.MessageTypeAPIRequest:
		actives := s.callAPIs(msgCtx.Context(), msgCtx)
		s.activeTentacle(actives)
	case message.MessageTypeState:
		// handle state message.
		watchKeys := s.msgHandler(msgCtx)
		// active tentacles.
		s.activeTentacle(watchKeys)
	default:
		log.Error("invalid message type", zfield.Header(msgCtx.Attributes()))
		return xerrors.ErrInvalidMessageType
	}
	return nil
}

func (s *statem) State() State {
	return State{
		ID:    s.ID,
		Props: s.Properties,
	}
}

type State struct {
	ID    string
	Props map[string]constraint.Node
}

func (s *State) Patch(op constraint.PatchOperator, path string, value []byte) (constraint.Node, error) {
	var (
		err    error
		result constraint.Node
	)
	if !strings.ContainsAny(path, ".[") {
		if result, err = s.patchProp(op, path, value); nil != err {
			log.Error("patch state property", zap.Error(err), zfield.Eid(s.ID))
		}
		return result, errors.Wrap(err, "patch state property")
	}

	// if path contains '.' or '[' .
	index := strings.IndexAny(path, ".[")
	propertyID, patchPath := path[:index], path[index:]

	valNode := constraint.NewNode(value)
	if result, err = constraint.Patch(s.Props[propertyID], valNode, patchPath, op); nil != err {
		log.Error("patch state", zfield.Path(path), zap.Error(err), zfield.Eid(s.ID))
		return nil, errors.Wrap(err, "patch state")
	}

	switch op {
	case constraint.PatchOpCopy:
		return result, nil
	}

	s.Props[propertyID] = result
	return result, nil
}

func (s *State) patchProp(op constraint.PatchOperator, path string, value []byte) (constraint.Node, error) {
	var (
		err    error
		result constraint.Node
	)
	switch op {
	case constraint.PatchOpReplace:
		s.Props[path] = constraint.NewNode(value)
	case constraint.PatchOpAdd:
		// patch property add.
		prop := s.Props[path]
		if nil == prop {
			prop = constraint.JSONNode(`[]`)
		}

		// patch add val.
		valNode := constraint.NewNode(value)
		if result, err = constraint.Patch(prop, valNode, "", op); nil != err {
			log.Error("patch add", zfield.Path(path), zap.Error(err))
			return result, errors.Wrap(err, "patch add")
		}
		s.Props[path] = result
	case constraint.PatchOpRemove:
		delete(s.Props, path)
	case constraint.PatchOpCopy:
		result = s.Props[path]
	default:
		return result, constraint.ErrJSONPatchReservedOp
	}
	return result, nil
}
