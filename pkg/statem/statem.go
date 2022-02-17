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
	"github.com/tkeel-io/core/pkg/environment"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const (
	// statem type enumerates.
	StateTypeBase = "sm-base"

	StateDisposingIdle  int32 = 0
	StateDisposingSync  int32 = 1
	StateDisposingAsync int32 = 2

	// statem runtime-status enumerates.
	StateDetached int32 = 0
	StateAttached int32 = 1

	// statem status enumerates.
	StateStatusActive   = "active"
	StateStatusInactive = "inactive"
	StateStatusDeleted  = "deleted"
)

func stateRuntimeStatusString(statusNum int32) string {
	switch statusNum {
	case StateDetached:
		return "detached"
	case StateAttached:
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
	ID           string                       `json:"id" msgpack:"id" mapstructure:"id"`
	Type         string                       `json:"type" msgpack:"type" mapstructure:"type"`
	Owner        string                       `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source       string                       `json:"source" msgpack:"source" mapstructure:"source"`
	Version      int64                        `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime     int64                        `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	Mappers      []MapperDesc                 `json:"mappers" msgpack:"mappers" mapstructure:"mappers"`
	KValues      map[string]constraint.Node   `json:"-" msgpack:"properties" mapstructure:"-"`
	Configs      map[string]constraint.Config `json:"configs" msgpack:"-" mapstructure:"-"`
	ConfigsBytes []byte                       `json:"-" msgpack:"configs_bytes" mapstructure:"-"`
	Properties   map[string]interface{}       `json:"properties" msgpack:"-" mapstructure:"-"`
}

func (b *Base) Copy() Base {
	bytes, _ := EncodeBase(b)
	bb, _ := DecodeBase(bytes)
	return *bb
}

func (b *Base) CopyPropertiesForJSON() {
	b.Properties = make(map[string]interface{})
	for k, v := range b.KValues {
		b.Properties[k] = v.Value()
	}
}

func (b *Base) DuplicateExpectValue() Base {
	cp := Base{
		ID:       b.ID,
		Type:     b.Type,
		Owner:    b.Owner,
		Source:   b.Source,
		Version:  b.Version,
		LastTime: b.LastTime,
		KValues:  make(map[string]constraint.Node),
		Configs:  make(map[string]constraint.Config),
	}

	cp.Mappers = append(cp.Mappers, b.Mappers...)
	return cp
}

func (b *Base) GetProperty(path string) (constraint.Node, error) {
	if !strings.ContainsAny(path, ".[") {
		if _, has := b.KValues[path]; !has {
			return constraint.NullNode{}, ErrPropertyNotFound
		}
		return b.KValues[path], nil
	}

	// patch copy property.
	arr := strings.SplitN(path, ".", 2)
	res, err := constraint.Patch(b.KValues[arr[0]], nil, arr[1], constraint.PatchOpCopy)
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
			return cfg, ErrPropertyNotFound
		}
		return b.Configs[segs[0]], nil
	}
	return cfg, errors.Wrap(constraint.ErrPatchPathInvalid, "copy config")
}

// statem state marchins.
type statem struct {
	Base

	// mapper & tentacles.
	mappers        map[string]mapper.Mapper              // key=mapperId
	tentacles      map[string][]mapper.Tentacler         // key=Sid#propertyKey
	cacheProps     map[string]map[string]constraint.Node // cache other property.
	indexTentacles map[string][]mapper.Tentacler         // key=targetId(mapperId/Sid)

	constraints        map[string]*constraint.Constraint
	searchConstraints  sort.StringSlice
	tseriesConstraints sort.StringSlice

	// mailbox & state runtime status.
	mailBox      *mailbox
	attached     int32
	disposing    int32
	nextFlushNum int32
	stateManager StateManager
	msgHandler   MessageHandler

	status Status

	ctx    context.Context
	cancel context.CancelFunc
}

// newEntity create an statem object.
func NewState(ctx context.Context, stateMgr StateManager, in *Base, msgHandler MessageHandler) (StateMachiner, error) {
	if in.ID == "" {
		in.ID = uuid()
	}

	ctx, cancel := context.WithCancel(ctx)

	state := &statem{
		Base: in.Copy(),

		ctx:            ctx,
		cancel:         cancel,
		stateManager:   stateMgr,
		msgHandler:     msgHandler,
		mailBox:        newMailbox(10),
		status:         SMStatusActive,
		disposing:      StateDisposingIdle,
		nextFlushNum:   StateFlushPeried,
		mappers:        make(map[string]mapper.Mapper),
		cacheProps:     make(map[string]map[string]constraint.Node),
		indexTentacles: make(map[string][]mapper.Tentacler),
		constraints:    make(map[string]*constraint.Constraint),
	}

	// initialize KValues.
	if nil == state.Base.KValues {
		state.KValues = make(map[string]constraint.Node)
	}

	// initialize Configs.
	if nil == state.Configs {
		state.Configs = make(map[string]constraint.Config)
	}

	// set KValues into cacheProps.
	state.cacheProps[in.ID] = state.KValues

	// use internal messga handler default.
	if nil == msgHandler {
		state.msgHandler = state.internelMessageHandler
	}

	return state, nil
}

func (s *statem) Setup() error {
	return nil
}

func (s *statem) GetID() string {
	return s.ID
}

func (s *statem) GetBase() *Base {
	return &s.Base
}

func (s *statem) GetStatus() Status {
	return s.status
}

func (s *statem) SetStatus(status Status) {
	s.status = status
}

func (s *statem) LoadEnvironments(env environment.ActorEnv) {
	s.tentacles = make(map[string][]mapper.Tentacler)

	// load actor mappers.
	for _, m := range env.Mappers {
		s.mappers[m.ID()] = m
		log.Debug("load environments, mapper ", logger.EntityID(s.ID), zap.String("TQL", m.String()))
	}

	var watchKeys []mapper.WatchKey
	// load actor tentacles.
	for _, t := range env.Tentacles {
		for _, item := range t.Items() {
			watchKeys = append(watchKeys, item)
			s.tentacles[item.String()] = append(s.tentacles[item.String()], t)
			log.Debug("load environments, watching ", logger.EntityID(s.ID), zap.String("WatchKey", item.String()))
		}
		log.Debug("load environments, tentacle ", logger.EntityID(s.ID), zap.String("tid", t.ID()), zap.String("target", t.TargetID()), zap.String("type", t.Type()), zap.Any("items", t.Items()))
	}

	s.flushState(context.Background())
	s.activeTentacle(watchKeys)
}

func (s *statem) GetManager() StateManager {
	return s.stateManager
}

// SetConfigs set entity configs.
func (s *statem) SetConfigs(configs map[string]constraint.Config) error {
	// reset state machine configs.
	s.Configs = make(map[string]constraint.Config)
	s.constraints = make(map[string]*constraint.Constraint)
	s.searchConstraints = make(sort.StringSlice, 0)
	s.tseriesConstraints = make(sort.StringSlice, 0)

	for key, cfg := range configs {
		s.Configs[key] = cfg
		if ct := constraint.NewConstraintsFrom(cfg); nil != ct {
			s.constraints[ct.ID] = ct
			// generate search indexes.
			if searchIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagSearch); len(searchIndexes) > 0 {
				s.searchConstraints = SliceAppend(s.searchConstraints, searchIndexes)
			}
			// generate time-series indexes.
			if tseriesIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagTimeSeries); len(tseriesIndexes) > 0 {
				s.tseriesConstraints = SliceAppend(s.tseriesConstraints, tseriesIndexes)
			}
		}
	}
	return nil
}

func (s *statem) getParent(segs []string) (*constraint.Config, int, error) {
	if len(segs) == 0 {
		return nil, 0, constraint.ErrPatchPathInvalid
	}

	// check patch path.
	for index, seg := range segs {
		if strings.TrimSpace(seg) == "" {
			return nil, index, constraint.ErrPatchPathInvalid
		}
	}

	var ok bool
	var cfg constraint.Config
	if cfg, ok = s.Configs[segs[0]]; !ok {
		return nil, 0, constraint.ErrPatchPathRoot
	}

	index, preCfg, err := cfg.GetConfig(segs, 1)
	return preCfg, index, errors.Wrap(err, "prev config not found")
}

func (s *statem) makePath(segs []string, cfg *constraint.Config) (cc constraint.Config, err error) {
	c := constraint.Config{
		ID:                segs[0],
		Type:              constraint.PropertyTypeStruct,
		Enabled:           true,
		EnabledSearch:     true,
		EnabledTimeSeries: true,
		Define:            make(map[string]interface{}),
		LastTime:          util.UnixMilli(),
	}

	if len(segs) > 1 {
		if cc, err = s.makePath(segs[1:], cfg); nil != err {
			return c, errors.Wrap(err, "make patch")
		} else if err = c.AppendField(cc); nil != err {
			return c, errors.Wrap(err, "make patch")
		}
	} else if err = c.AppendField(*cfg); nil != err {
		return c, errors.Wrap(err, "make patch")
	}

	return c, nil
}

// PatchConfigs set entity configs.
func (s *statem) PatchConfigs(patchData []*PatchData) error { //nolint
	for _, pd := range patchData {
		var segment string
		cfg, _ := pd.Value.(constraint.Config)
		segs := strings.Split(strings.TrimSpace(pd.Path), ".")

		// set values.
		cfg.ID = segs[len(segs)-1]
		cfg.LastTime = util.UnixMilli()

		if len(segs) > 1 {
			segment = segs[len(segs)-1]
			parentCfg, index, err := s.getParent(segs[:len(segs)-1])
			if errors.Is(err, constraint.ErrPatchPathRoot) {
				segment = segs[0]
				if cfg, err = s.makePath(segs[:len(segs)-1], &cfg); nil != err {
					log.Error("make patch path",
						zap.Error(err),
						logger.EntityID(s.ID),
						zap.Any("config", pd.Value),
						zap.String("path", pd.Path))
					return errors.Wrap(err, "make patch path")
				}
			} else if errors.Is(err, constraint.ErrPatchPathLack) {
				segment = segs[index]
				if cfg, err = s.makePath(segs[index:len(segs)-1], &cfg); nil != err {
					log.Error("make patch path",
						zap.Error(err),
						logger.EntityID(s.ID),
						zap.Any("config", pd.Value),
						zap.String("path", pd.Path))
					return errors.Wrap(err, "make patch path")
				}
			} else if nil != err {
				log.Error("get parent config",
					zap.Error(err),
					logger.EntityID(s.ID),
					zap.Any("config", pd.Value),
					zap.String("path", pd.Path))
				return errors.Wrap(err, "state machine patch configs")
			}

			log.Debug("patch state machine configs", logger.EntityID(s.ID), zap.Strings("segments", segs), zap.Any("value", cfg), zap.Int("index", index))

			switch pd.Operator {
			case constraint.PatchOpAdd:
				fallthrough
			case constraint.PatchOpReplace:
				if index == 0 {
					s.Configs[segment] = cfg
				} else if err = parentCfg.AppendField(cfg); nil != err {
					return errors.Wrap(err, "upsert state machine configs")
				}
			case constraint.PatchOpRemove:
				if index == len(segs)-1 || nil != parentCfg {
					if err = parentCfg.RemoveField(segment); nil != err {
						return errors.Wrap(err, "remove state machine configs")
					}
				}
			case constraint.PatchOpCopy:
				// TODO: 在这里处理的时候，sync-actor-loop, 以消息的形式，返回值是难处理的.
			}
		} else if len(segs) == 1 {
			switch pd.Operator {
			case constraint.PatchOpAdd:
				fallthrough
			case constraint.PatchOpReplace:
				s.Configs[cfg.ID] = cfg
			case constraint.PatchOpRemove:
				delete(s.Configs, segs[0])
			}
		}
		log.Debug("patch state machine config", zap.String("path", pd.Path), zap.Any("value", pd.Value))
	}

	s.constraints = make(map[string]*constraint.Constraint)
	s.searchConstraints = make(sort.StringSlice, 0)
	s.tseriesConstraints = make(sort.StringSlice, 0)

	for key, cfg := range s.Configs {
		s.Configs[key] = cfg
		if ct := constraint.NewConstraintsFrom(cfg); nil != ct {
			s.constraints[ct.ID] = ct
			// generate search indexes.
			if searchIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagSearch); len(searchIndexes) > 0 {
				s.searchConstraints = SliceAppend(s.searchConstraints, searchIndexes)
			}
			// generate time-series indexes.
			if tseriesIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagTimeSeries); len(tseriesIndexes) > 0 {
				s.tseriesConstraints = SliceAppend(s.tseriesConstraints, tseriesIndexes)
			}
		}
	}

	log.Debug("patch state machine configs", zap.Any("value", patchData))

	return nil
}

// AppendConfigs append entity configs.
func (s *statem) AppendConfigs(configs map[string]constraint.Config) error {
	for key, cfg := range configs {
		s.Configs[key] = cfg
		if ct := constraint.NewConstraintsFrom(cfg); nil != ct {
			s.constraints[ct.ID] = ct
			// generate search indexes.
			if searchIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagSearch); len(searchIndexes) > 0 {
				s.searchConstraints = SliceAppend(s.searchConstraints, searchIndexes)
			}
			// generate time-series indexes.
			if tseriesIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagTimeSeries); len(tseriesIndexes) > 0 {
				s.tseriesConstraints = SliceAppend(s.tseriesConstraints, tseriesIndexes)
			}
		}
	}
	return nil
}

// RemoveConfigs remove entity property config.
func (s *statem) RemoveConfigs(propertyIDs []string) error {
	// delete property config.
	for _, propertyID := range propertyIDs {
		delete(s.Configs, propertyID)
		delete(s.constraints, propertyID)
	}

	// reset indexes.
	s.searchConstraints = make(sort.StringSlice, 0)
	s.tseriesConstraints = make(sort.StringSlice, 0)

	// reparse property configs.
	for _, ct := range s.constraints {
		// generate search indexes.
		if searchIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagSearch); len(searchIndexes) > 0 {
			s.searchConstraints = SliceAppend(s.searchConstraints, searchIndexes)
		}
		// generate time-series indexes.
		if tseriesIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagTimeSeries); len(tseriesIndexes) > 0 {
			s.tseriesConstraints = SliceAppend(s.tseriesConstraints, tseriesIndexes)
		}
	}
	return nil
}

func (s *statem) SetMessageHandler(msgHandler MessageHandler) {
	s.msgHandler = msgHandler
}

// OnMessage recive statem input messages.
func (s *statem) OnMessage(message Message) bool {
	var (
		reqID     = uuid()
		attaching = false
	)

	if s.status == SMStatusDeleted {
		return false
	}

	log.Debug("statem.OnMessage", logger.EntityID(s.ID), logger.RequestID(reqID))

	switch msg := message.(type) {
	case TentacleMsg:
		s.invokeTentacleMsg(msg)
	default:
		// dispose message.
		watchKeys := s.msgHandler(message)
		// active tentacles.
		s.activeTentacle(watchKeys)
	}

	message.Promised(s)
	s.flush(context.Background())

	return attaching
}

// InvokeMsg run loopHandler.
func (s *statem) HandleLoop() {
	var (
		Ensure  = 3
		message Message
	)

	for {
		if s.nextFlushNum == 0 {
			// flush properties.
			if err := s.flush(s.ctx); nil != err {
				log.Error("flush state properties", logger.EntityID(s.ID), zap.Error(err))
			}
		}

		// consume message from mailbox.
		if message = s.mailBox.Get(); nil == message {
			if Ensure > 0 {
				Ensure--
				runtime.Gosched()
				continue
			}

			// detach this statem.
			if !atomic.CompareAndSwapInt32(&s.attached, StateAttached, StateDetached) {
				log.Error("exception occurred, mismatched statem runtime status.",
					logger.Status(stateRuntimeStatusString(atomic.LoadInt32(&s.attached))))
			}

			// flush properties.
			if err := s.flush(s.ctx); nil != err {
				log.Error("flush state properties failed", logger.EntityID(s.ID), zap.Error(err))
			}
			// detach coroutins.
			break
		}

		// reset be surs.
		Ensure = 3
		s.nextFlushNum = (s.nextFlushNum + StateFlushPeried - 1) % StateFlushPeried
	}

	log.Info("detached statem.", logger.EntityID(s.ID))
}

// InvokeMsg dispose statem input messages.
func (s *statem) internelMessageHandler(message Message) []WatchKey {
	switch msg := message.(type) {
	case PropertyMessage:
		return s.invokePropertyMsg(msg)
	default:
		// invalid msg typs.
		log.Error("undefine message type", logger.EntityID(s.ID), logger.MessageInst(msg))
	}

	return nil
}

// invokePropertyMsg dispose Property messags.
func (s *statem) invokePropertyMsg(msg PropertyMessage) []WatchKey {
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
		if s.ID == msg.StateID {
			if err := s.setProperty(constraint.NewPatchOperator(msg.Operator), key, value); nil != err {
				log.Error("set entity property failed ", logger.EntityID(s.ID), logger.PropertyKey(key), zap.Error(err))
			}
		} else {
			stateProps[key] = value
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
					val, err := s.getProperty(thisStateProps, active.PropertyKey)
					if nil != err {
						log.Warn("get property", zap.Error(err), logger.EntityID(s.ID),
							zap.String("entity", active.EntityId), zap.String("property-key", active.PropertyKey))
						continue
					}
					messages[targetID][active.PropertyKey] = val
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
		if stateID == s.ID {
			log.Warn("send message to self", logger.EntityID(s.ID))
			continue
		}
		s.stateManager.SendMsg(MessageContext{
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
func (s *statem) activeMapper(actives map[string][]mapper.Tentacler) {
	if len(actives) == 0 {
		return
	}

	var err error
	var activeKeys []mapper.WatchKey
	for mapperID := range actives {
		input := make(map[string]constraint.Node)
		for _, tentacle := range s.mappers[mapperID].Tentacles() {
			if tentacle.Type() == mapper.TentacleTypeMapper {
				for _, item := range tentacle.Items() {
					var val constraint.Node
					if s.ID == item.EntityId {
						if val, err = s.getProperty(s.cacheProps[item.EntityId], item.PropertyKey); nil != err {
							log.Error("get property failed", logger.RequestID(item.PropertyKey), zap.Error(err))
							continue
						}
					} else {
						if props, ok := s.cacheProps[item.EntityId]; ok {
							val = props[item.PropertyKey]
						}
					}
					if nil != val {
						input[item.String()] = val
					}
				}
			}
		}

		if len(input) == 0 {
			log.Debug("obtain mapper input, empty params", logger.MapperID(mapperID))
			continue
		}

		var properties map[string]constraint.Node

		// excute mapper.
		if properties, err = s.mappers[mapperID].Exec(input); nil != err {
			log.Error("exec statem mapper failed ", zap.Error(err))
		}

		log.Debug("exec mapper", logger.MapperID(mapperID), zap.Any("input", input), zap.Any("output", properties))

		if len(properties) > 0 {
			for propertyKey, value := range properties {
				if err = s.setProperty(constraint.PatchOpReplace, propertyKey, value); nil != err {
					log.Error("set property failed", logger.EntityID(s.ID),
						zap.String("property_key", propertyKey), zap.Error(err))
					continue
				}
				s.LastTime = time.Now().UnixNano() / 1e6
				activeKeys = append(activeKeys, mapper.WatchKey{EntityId: s.ID, PropertyKey: propertyKey})
			}
		}
	}

	s.activeTentacle(unique(activeKeys))
}

func unique(actives []mapper.WatchKey) []mapper.WatchKey {
	umap := make(map[string]mapper.WatchKey)
	for _, w := range actives {
		umap[w.String()] = w
	}

	actives = []mapper.WatchKey{}
	for _, w := range umap {
		actives = append(actives, w)
	}
	return actives
}

func (s *statem) getProperty(properties map[string]constraint.Node, propertyKey string) (constraint.Node, error) {
	if !strings.ContainsAny(propertyKey, ".[") {
		if _, has := properties[propertyKey]; !has {
			return constraint.NullNode{}, ErrPropertyNotFound
		}
		return properties[propertyKey], nil
	}

	// patch property.
	arr := strings.SplitN(propertyKey, ".", 2)
	res, err := constraint.Patch(properties[arr[0]], nil, arr[1], constraint.PatchOpCopy)
	return res, errors.Wrap(err, "get patch failed")
}

func (s *statem) setProperty(op constraint.PatchOperator, propertyKey string, value constraint.Node) error {
	var err error
	var resultNode constraint.Node

	if !strings.ContainsAny(propertyKey, ".[") {
		switch op {
		case constraint.PatchOpReplace:
			s.KValues[propertyKey] = value
		case constraint.PatchOpAdd:
			// patch property add.
			val := s.KValues[propertyKey]
			if nil == val {
				val = constraint.JSONNode(`[]`)
			}
			if resultNode, err = constraint.Patch(val, value, "", op); nil != err {
				log.Error("set property failed", logger.EntityID(s.ID), zap.Error(err))
				return errors.Wrap(err, "set property failed")
			}
			s.KValues[propertyKey] = resultNode
		case constraint.PatchOpRemove:
			delete(s.KValues, propertyKey)
		default:
			return constraint.ErrJSONPatchReservedOp
		}
		return nil
	}

	// patch property.
	index := strings.IndexAny(propertyKey, ".[")
	propertyID, patchPath := propertyKey[:index], propertyKey[index:]
	if _, has := s.KValues[propertyID]; !has {
		return constraint.ErrPatchNotFound
	}

	if resultNode, err = constraint.Patch(s.KValues[propertyID], value, patchPath, op); nil != err {
		log.Error("set property failed", logger.EntityID(s.ID), zap.Error(err))
		return errors.Wrap(err, "set property failed")
	}

	s.KValues[propertyID] = resultNode
	return nil
}

// invokeTentacleMsg dispose Tentacle messages.
func (s *statem) invokeTentacleMsg(msg TentacleMsg) {
	if s.ID == msg.StateID {
		// ignore this messags.
		return
	}

	switch msg.Operator {
	case TentacleOperatorAppend:
		tentacle := mapper.NewRemoteTentacle(mapper.TentacleTypeEntity, msg.StateID, msg.Items)
		s.indexTentacles[msg.StateID] = []mapper.Tentacler{tentacle}
	case TentacleOperatorRemove:
		delete(s.indexTentacles, msg.StateID)
	default:
		log.Error("invalid tentacle operator",
			logger.Operator(msg.Operator), logger.MessageInst(msg))
	}

	log.Debug("catch tentacle event.",
		logger.Operator(msg.Operator), logger.EntityID(msg.StateID), logger.MessageInst(msg))

	// generate tentacles again.
	s.generateTentacles()
}

func (s *statem) generateTentacles() {
	s.tentacles = make(map[string][]mapper.Tentacler)
	for _, tentacles := range s.indexTentacles {
		for _, tentacle := range tentacles {
			if mapper.TentacleTypeMapper == tentacle.Type() || tentacle.IsRemote() {
				log.Debug("setup tentacle",
					logger.EntityID(s.ID),
					logger.Type(tentacle.Type()),
					logger.Target(tentacle.TargetID()))
				for _, item := range tentacle.Items() {
					s.tentacles[item.String()] = append(s.tentacles[item.String()], tentacle)
				}
			}
		}
	}
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
