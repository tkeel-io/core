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
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/types"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
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
	cacheProps map[string]map[string]tdtl.Node // cache other property.

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

	// state machine message handler.
	republisher MessageHandler

	// state Context, state context version.
	sCtx    StateContext
	version int64

	ctx    context.Context
	cancel context.CancelFunc
}

// NewState create an statem object.
func NewState(ctx context.Context, in *dao.Entity, dispatcher dispatch.Dispatcher, resourceManager types.ResourceManager, republisher MessageHandler) (Machiner, error) {
	ctx, cancel := context.WithCancel(ctx)

	state := &statem{
		Entity: in.Copy(),

		ctx:             ctx,
		cancel:          cancel,
		republisher:     republisher,
		dispatcher:      dispatcher,
		resourceManager: resourceManager,
		mappers:         make(map[string]mapper.Mapper),
		cacheProps:      make(map[string]map[string]tdtl.Node),
		constraints:     make(map[string]*constraint.Constraint),
	}

	// initialize Properties.
	if nil == state.Entity.Properties {
		state.Properties = make(map[string]tdtl.Node)
	}

	if nil == republisher {
		// default message handler.
		state.republisher = state.invokeRepublishMessage
	}

	// initialize state context.
	state.sCtx = newContext(state)

	return state, nil
}

// GetID returns state ID.
func (s *statem) GetID() string {
	return s.ID
}

func (s *statem) GetEntity() *dao.Entity {
	return &s.Entity
}

func (s *statem) Context() *StateContext {
	return &s.sCtx
}

func (s *statem) updateFromContext() {
	if atomic.LoadInt64(&s.sCtx.version) > s.version {
		log.Debug("state context changed", zfield.Eid(s.ID), zfield.Type(s.Type))
		s.mappers = s.sCtx.mappers
		s.tentacles = make(map[string][]mapper.Tentacler)

		// deploy tentacles.
		for _, t := range s.sCtx.tentacles {
			for _, item := range t.Items() {
				s.tentacles[item.String()] = append(s.tentacles[item.String()], t)
				log.Debug("load environments, watching ", zfield.Eid(s.ID), zap.String("WatchKey", item.String()))
			}
			log.Debug("load environments, tentacle ", zfield.Eid(s.ID), zap.String("tid", t.ID()), zap.String("target", t.TargetID()), zap.String("type", t.Type()), zap.Any("items", t.Items()))
		}

		// set version.
		atomic.SwapInt64(&s.version, s.sCtx.version)
	}
}

// OnMessage recive statem input messages.
func (s *statem) Invoke(ctx context.Context, msgCtx message.Context) Result {
	// update state from StateContext.
	s.updateFromContext()

	// delive message.
	var actives []WatchKey
	result := Result{Status: MCompleted}
	msgType := msgCtx.Get(message.ExtMessageType)
	switch message.MessageType(msgType) {
	case message.MessageTypeRaw:
		actives = s.invokeRawMessage(ctx, msgCtx)
	case message.MessageTypeState:
		actives = s.invokeStateMessage(ctx, msgCtx)
	case message.MessageTypeAPIRepublish:
		actives = s.republisher(ctx, msgCtx)
	case message.MessageTypeAPIRequest:
		actives, result = s.callAPIs(ctx, msgCtx)
		s.activeTentacle(actives)
		return result
	case message.MessageTypeMapperInit:
		actives = s.invokeMapperInit(ctx, msgCtx)
	default:
		log.Error("message type not support", zfield.Type(msgType))
	}

	// exec tentacles.
	if result.Err == nil {
		s.activeTentacle(actives)
		s.flush(ctx)
	}

	return result
}

func (s *statem) State() State {
	return State{
		ID:    s.ID,
		Props: s.Properties,
	}
}

type State struct {
	ID    string
	Props map[string]tdtl.Node
}

func (s *State) Get(path string) (tdtl.Node, error) {
	val, err := s.Patch(xjson.OpCopy, path, nil)
	return val, errors.Wrap(err, "patch copy property")
}

func (s *State) Patch(op xjson.PatchOp, path string, value []byte) (tdtl.Node, error) {
	var (
		err    error
		result tdtl.Node
	)
	if !strings.ContainsAny(path, ".[") {
		result, err = s.patchProp(op, path, string(value))
		return result, errors.Wrap(err, "patch state property")
	}

	// if path contains '.' or '[' .
	index := strings.IndexAny(path, ".[")
	propertyID, patchPath := path[:index], strings.TrimPrefix(path[index:], ".")

	valNode := tdtl.New(value)
	if result, err = xjson.Patch(s.get(propertyID), valNode, patchPath, op); nil != err {
		return nil, errors.Wrap(err, "patch state")
	}

	switch op {
	case xjson.OpCopy:
		return result, nil
	}

	s.Props[propertyID] = result
	return result, nil
}

func (s *State) get(pid string) tdtl.Node {
	if val, ok := s.Props[pid]; ok {
		return val
	}
	return tdtl.New("")
}

func (s *State) patchProp(op xjson.PatchOp, path string, value string) (tdtl.Node, error) {
	var (
		err    error
		bytes  []byte
		result tdtl.Node
	)
	switch op {
	case xjson.OpReplace:
		s.Props[path] = tdtl.New(value)
	case xjson.OpAdd:
		// patch property add.
		prop := s.Props[path]
		if nil == prop {
			prop = tdtl.New(`[]`)
		}

		// patch add val.
		bytes, err = collectjs.Append([]byte(prop.String()), path, []byte(value))
		if nil != err {
			return result, errors.Wrap(err, "patch add")
		}
		result = tdtl.New(bytes)
		s.Props[path] = result
	case xjson.OpRemove:
		delete(s.Props, path)
	case xjson.OpCopy:
		var ok bool
		if result, ok = s.Props[path]; !ok {
			return result, xerrors.ErrPropertyNotFound
		}
	default:
		return result, xerrors.ErrJSONPatchReservedOp
	}
	return result, nil
}
