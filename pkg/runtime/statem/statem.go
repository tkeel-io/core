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

package statem

// statem: state marchins.

import (
	"context"
	"crypto/rand"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	cerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const (
	// statem runtime-status enumerates.
	StateRuntimeDetached int32 = 0
	StateRuntimeAttached int32 = 1

	// state machine default configurations.
	defaultEnsureConsumeTimes int = 3
	defaultStateFlushPeried   int = 10

	// statem status enumerates.
	SMStatusActive   Status = "active"
	SMStatusInactive Status = "inactive"
	SMStatusDeleted  Status = "deleted"
)

// statem state marchins.
type statem struct {
	// state basic fields.
	Base
	// other state machine property cache.
	cacheProps map[string]map[string]constraint.Node // cache other property.

	// mapper & tentacles.
	mappers   map[string]mapper.Mapper      // key=mapperId
	tentacles map[string][]mapper.Tentacler // key=Sid#propertyKey

	// parse from Configs.
	constraints        map[string]*constraint.Constraint
	searchConstraints  sort.StringSlice
	tseriesConstraints sort.StringSlice

	// state machine mailbox.
	mailbox *mailbox
	// state manager.
	stateManager StateManager

	status             Status
	attached           int32
	nextFlushNum       int
	ensureComsumeTimes int
	// state machine message handler.
	msgHandler MessageHandler

	sCtx   StateContext
	ctx    context.Context
	cancel context.CancelFunc
}

// newEntity create an statem object.
func NewState(ctx context.Context, stateManager StateManager, in *Base, msgHandler MessageHandler) (StateMachiner, error) {
	if in.ID == "" {
		in.ID = uuid()
	}

	ctx, cancel := context.WithCancel(ctx)

	state := &statem{
		Base: in.Copy(),

		ctx:                ctx,
		cancel:             cancel,
		mailbox:            newMailbox(20),
		status:             SMStatusActive,
		msgHandler:         msgHandler,
		stateManager:       stateManager,
		nextFlushNum:       defaultStateFlushPeried,
		ensureComsumeTimes: defaultEnsureConsumeTimes,
		mappers:            make(map[string]mapper.Mapper),
		cacheProps:         make(map[string]map[string]constraint.Node),
		constraints:        make(map[string]*constraint.Constraint),
	}

	// initialize Properties.
	if nil == state.Base.Properties {
		state.Properties = make(map[string]constraint.Node)
	}

	// initialize Configs.
	if nil == state.Configs {
		state.Configs = make(map[string]constraint.Config)
	}

	// set Properties into cacheProps.
	state.cacheProps[in.ID] = state.Properties

	// use internal messga handler default.
	if nil == msgHandler {
		state.msgHandler = state.internelMessageHandler
	}

	return state, nil
}

// GetID returns state ID.
func (s *statem) GetID() string {
	return s.ID
}

// GetBase return state basic info.
func (s *statem) GetBase() *Base {
	return &s.Base
}

// GetStatus returns state machine status.
func (s *statem) GetStatus() Status {
	return s.status
}

// WithContext set state Context.
func (s *statem) WithContext(ctx StateContext) StateMachiner {
	s.sCtx = ctx
	return s
}

// OnMessage recive statem input messages.
func (s *statem) OnMessage(msg Message) bool {
	var attaching = false
	if s.status == SMStatusDeleted {
		return false
	}

	for {
		if nil == s.mailbox.Put(msg) {
			break
		}
		runtime.Gosched()
	}

	if atomic.CompareAndSwapInt32(&s.attached, StateRuntimeDetached, StateRuntimeAttached) {
		attaching = true
	}

	return attaching
}

// HandleLoop run loopHandler.
func (s *statem) HandleLoop() {
	var message Message
	var ensureComsumeTimes = s.ensureComsumeTimes
	log.Debug("actor attached", zfield.ID(s.ID))

	for {
		if s.nextFlushNum == 0 {
			// flush properties.
			if err := s.flush(s.ctx); nil != err {
				log.Error("flush state properties", zfield.ID(s.ID), zap.Error(err))
			}
		}

		// consume message from mailbox.
		if s.mailbox.Empty() {
			if ensureComsumeTimes > 0 {
				ensureComsumeTimes--
				runtime.Gosched()
				continue
			}

			// detach this statem.
			if !atomic.CompareAndSwapInt32(&s.attached, StateRuntimeAttached, StateRuntimeDetached) {
				log.Error("exception occurred, mismatched statem runtime status.",
					zfield.Status(stateRuntimeStatusString(atomic.LoadInt32(&s.attached))))
			}

			// flush properties.
			if err := s.flush(s.ctx); nil != err {
				log.Error("flush state properties failed", zfield.ID(s.ID), zap.Error(err))
			}
			// detach coroutins.
			break
		}

		message = s.mailbox.Get()
		switch msg := message.(type) {
		default:
			// handle message.
			watchKeys := s.msgHandler(msg)
			// active tentacles.
			s.activeTentacle(watchKeys)
		}

		message.Promised(s)

		// reset be sure.
		ensureComsumeTimes = s.ensureComsumeTimes
		s.nextFlushNum = (s.nextFlushNum + defaultStateFlushPeried - 1) % defaultStateFlushPeried
	}

	log.Info("detached statem.", zfield.ID(s.ID))
}

// internelMessageHandler dispose statem input messages.
func (s *statem) internelMessageHandler(message Message) []WatchKey {
	switch msg := message.(type) {
	case PropertyMessage:
		return s.invokePropertyMessage(msg)
	default:
		// invalid msg typs.
		log.Error("undefine message type", zfield.ID(s.ID), zfield.MessageInst(msg))
	}

	return nil
}

// invokePropertyMessage invoke property message.
func (s *statem) invokePropertyMessage(msg PropertyMessage) []WatchKey {
	setStateID := msg.StateID
	if setStateID == "" {
		setStateID = s.ID
	}

	watchKeys := make([]mapper.WatchKey, 0)
	if _, has := s.cacheProps[setStateID]; !has {
		s.cacheProps[setStateID] = make(map[string]constraint.Node)
	}

	stateProps := s.cacheProps[setStateID]
	for key, value := range msg.Properties {
		if _, err := patchProperty(stateProps, key, constraint.PatchOpReplace, value); nil != err {
			log.Error("set state property", zfield.ID(s.ID), zfield.PropertyKey(key), zap.Error(err))
			continue
		}
		watchKeys = append(watchKeys, mapper.WatchKey{EntityId: setStateID, PropertyKey: key})
	}

	// set last active tims.
	if setStateID == s.ID {
		s.Version++
		s.LastTime = time.Now().UnixNano() / 1e6
	}

	return watchKeys
}

// activeTentacle active tentacles.
func (s *statem) activeTentacle(actives []mapper.WatchKey) { //nolint
	if len(actives) == 0 {
		return
	}

	var (
		messages        = make(map[string]map[string]constraint.Node)
		activeTentacles = make(map[string][]mapper.Tentacler)
	)

	thisStateProps := s.cacheProps[s.ID]
	for _, active := range actives {
		// full match.
		if tentacles, exists := s.tentacles[active.String()]; exists {
			for _, tentacle := range tentacles {
				targetID := tentacle.TargetID()
				if mapper.TentacleTypeMapper == tentacle.Type() {
					activeTentacles[targetID] = append(activeTentacles[targetID], tentacle)
				} else if mapper.TentacleTypeEntity == tentacle.Type() {
					// make if not exists.
					if _, exists := messages[targetID]; !exists {
						messages[targetID] = make(map[string]constraint.Node)
					}

					// 在组装成Msg后，SendMsg的时候会对消息进行序列化，所以这里不需要Deep Copy.
					// 在这里我们需要解析PropertyKey, PropertyKey中可能存在嵌套层次.
					messages[targetID][active.PropertyKey] = thisStateProps[active.PropertyKey]
				} else {
					// undefined tentacle typs.
					log.Warn("undefined tentacle type", zap.Any("tentacle", tentacle))
				}
			}
		} else {
			// TODO...
			// 如果消息是缓存，那么，我们应该对改state的tentacles刷新。
			log.Debug("match end of string \".*\" PropertyKey.", zap.String("entity", active.EntityId), zap.String("property-key", active.PropertyKey))
			// match entityID.*   .
			for watchKey, tentacles := range s.tentacles {
				arr := strings.Split(watchKey, ".")
				if len(arr) == 2 && arr[1] == "*" && arr[0] == active.EntityId {
					for _, tentacle := range tentacles {
						targetID := tentacle.TargetID()
						if mapper.TentacleTypeMapper == tentacle.Type() {
							activeTentacles[targetID] = append(activeTentacles[targetID], tentacle)
						} else if mapper.TentacleTypeEntity == tentacle.Type() {
							// make if not exists.
							if _, exists := messages[targetID]; !exists {
								messages[targetID] = make(map[string]constraint.Node)
							}

							segments := strings.Split(active.PropertyKey, ".")
							// 在组装成Msg后，SendMsg的时候会对消息进行序列化，所以这里不需要Deep Copy.
							// 在这里我们需要解析PropertyKey, PropertyKey中可能存在嵌套层次.
							messages[targetID][segments[0]] = thisStateProps[segments[0]]
						} else {
							// undefined tentacle typs.
							log.Warn("undefined tentacle type", zap.Any("tentacle", tentacle))
						}
					}
				}
			}
		}
	}

	for stateID, msg := range messages {
		s.stateManager.RouteMessage(context.Background(),
			MessageContext{
				Headers: Header{
					MessageCtxHeaderSourceID: s.ID,
					MessageCtxHeaderTargetID: stateID,
				},
				Message: PropertyMessage{
					StateID:    s.ID,
					Properties: msg,
				},
			})
	}

	// active mapper.
	s.activeMapper(activeTentacles)
}

// activeMapper active mappers.
func (s *statem) activeMapper(actives map[string][]mapper.Tentacler) { //nolint
	if len(actives) == 0 {
		return
	}

	var err error
	for mapperID := range actives {
		input := make(map[string]constraint.Node)
		for _, tentacle := range s.mappers[mapperID].Tentacles() {
			for _, item := range tentacle.Items() {
				var val constraint.Node
				if val, err = s.getProperty(s.cacheProps[item.EntityId], item.PropertyKey); nil != err {
					log.Error("patch copy", zfield.RequestID(item.PropertyKey), zap.Error(err))
					continue
				} else if nil != val {
					input[item.String()] = val
				}
			}
		}

		if len(input) == 0 {
			log.Debug("obtain mapper input, empty params", zfield.MapperID(mapperID))
			continue
		}

		var properties map[string]constraint.Node

		// excute mapper.
		if properties, err = s.mappers[mapperID].Exec(input); nil != err {
			log.Error("exec statem mapper failed ", zap.Error(err))
		}

		log.Debug("exec mapper", zfield.MapperID(mapperID), zap.Any("input", input), zap.Any("output", properties))

		for propertyKey, value := range properties {
			if err = s.setProperty(propertyKey, constraint.PatchOpReplace, value); nil != err {
				log.Error("set property", zfield.ID(s.ID), zap.String("property_key", propertyKey), zap.Error(err))
				continue
			}
			s.LastTime = time.Now().UnixNano() / 1e6
		}
	}
}

func (s *statem) getProperty(properties map[string]constraint.Node, propertyKey string) (constraint.Node, error) {
	val, err := patchProperty(properties, propertyKey, constraint.PatchOpCopy, nil)
	return val, errors.Wrap(err, "patch copy property")
}

func (s *statem) setProperty(path string, op constraint.PatchOperator, value constraint.Node) error {
	_, err := patchProperty(s.Properties, path, constraint.PatchOpReplace, value)
	return errors.Wrap(err, "set property")
}

func patchProperty(props map[string]constraint.Node, path string, op constraint.PatchOperator, val constraint.Node) (constraint.Node, error) { //nolint
	var err error
	var resultNode constraint.Node
	if !strings.ContainsAny(path, ".[") {
		switch op {
		case constraint.PatchOpReplace:
			props[path] = val
		case constraint.PatchOpAdd:
			// patch property add.
			prop := props[path]
			if nil == prop {
				prop = constraint.JSONNode(`[]`)
			}

			// patch add val.
			if resultNode, err = constraint.Patch(val, prop, "", op); nil != err {
				log.Error("patch add", zfield.Path(path), zap.Error(err))
				return nil, errors.Wrap(err, "patch add")
			}
			props[path] = resultNode
		case constraint.PatchOpRemove:
			delete(props, path)
		case constraint.PatchOpCopy:
			resultNode = props[path]
		default:
			return nil, constraint.ErrJSONPatchReservedOp
		}
		return resultNode, nil
	}

	// if path contains '.' or '[' .
	index := strings.IndexAny(path, ".[")
	propertyID, patchPath := path[:index], path[index:]
	if _, has := props[propertyID]; !has {
		log.Error("patch state", zfield.Path(path), zap.Error(constraint.ErrPatchNotFound))
		return nil, constraint.ErrPatchNotFound
	}

	if resultNode, err = constraint.Patch(props[propertyID], val, patchPath, op); nil != err {
		log.Error("patch state", zfield.Path(path), zap.Error(err))
		return nil, errors.Wrap(err, "patch state")
	}

	props[propertyID] = resultNode
	return resultNode, nil
}

// uuid generate an uuid.
func uuid() string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return ""
	}
	// see section 4.1.1.
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// see section 4.1.3.
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func stateRuntimeStatusString(statusNum int32) string {
	switch statusNum {
	case StateRuntimeDetached:
		return "detached"
	case StateRuntimeAttached:
		return "attached"
	default:
		return "undefine"
	}
}

type MapperDesc struct {
	Name      string `json:"name"`
	TQLString string `json:"tql"` //nolint
}

// EntityBase statem basic informatinon.
type Base struct {
	ID         string                       `json:"id" msgpack:"id" mapstructure:"id"`
	Type       string                       `json:"type" msgpack:"type" mapstructure:"type"`
	Owner      string                       `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source     string                       `json:"source" msgpack:"source" mapstructure:"source"`
	Version    int64                        `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime   int64                        `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	Mappers    []MapperDesc                 `json:"mappers" msgpack:"mappers" mapstructure:"mappers"`
	Properties map[string]constraint.Node   `json:"properties" msgpack:"properties" mapstructure:"-"`
	Configs    map[string]constraint.Config `json:"configs" msgpack:"-" mapstructure:"-"`
	ConfigFile []byte                       `json:"-" msgpack:"config_file" mapstructure:"-"`
}

func (b *Base) Copy() Base {
	bytes, _ := EncodeBase(b)
	bb, _ := DecodeBase(bytes)
	return *bb
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
			return constraint.NullNode{}, cerrors.ErrPropertyNotFound
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
			return cfg, cerrors.ErrPropertyNotFound
		}
		return b.Configs[segs[0]], nil
	}
	return cfg, errors.Wrap(constraint.ErrPatchPathInvalid, "copy config")
}
