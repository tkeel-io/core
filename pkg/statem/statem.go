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
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
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
	ID       string                       `json:"id" msgpack:"id" mapstructure:"id"`
	Type     string                       `json:"type" msgpack:"type" mapstructure:"type"`
	Owner    string                       `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source   string                       `json:"source" msgpack:"source" mapstructure:"source"`
	Version  int64                        `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime int64                        `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	Mappers  []MapperDesc                 `json:"mappers" msgpack:"mappers" mapstructure:"mappers"`
	KValues  map[string]constraint.Node   `json:"properties" msgpack:"properties" mapstructure:"-"` //nolint
	Configs  map[string]constraint.Config `json:"configs" msgpack:"configs" mapstructure:"configs"`
}

func (b *Base) Copy() Base {
	bytes, _ := EncodeBase(b)
	bb, _ := DecodeBase(bytes)
	return *bb
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
func NewState(ctx context.Context, stateMgr StateManager, in *Base, msgHandler MessageHandler) (StateMarchiner, error) {
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
	descs := s.Mappers
	s.Mappers = make([]MapperDesc, 0)
	for _, desc := range descs {
		s.appendMapper(desc)
	}
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

func (s *statem) GetManager() StateManager {
	return s.stateManager
}

// SetConfigs set entity configs.
func (s *statem) SetConfigs(configs map[string]constraint.Config) error {
	for key, cfg := range configs {
		s.Configs[key] = cfg
		if ct := constraint.NewConstraintsFrom(cfg); nil != ct {
			s.constraints[ct.ID] = ct
			// generate search indexes.
			searchIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagSearch)
			if len(searchIndexes) > 0 {
				s.searchConstraints =
					SliceAppend(s.searchConstraints, searchIndexes)
			}
			// generate time-series indexes.
			tseriesIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagTimeSeries)
			if len(tseriesIndexes) > 0 {
				s.tseriesConstraints =
					SliceAppend(s.tseriesConstraints, tseriesIndexes)
			}
		}
	}
	return nil
}

// AppendConfigs append entity configs.
func (s *statem) AppendConfigs(configs map[string]constraint.Config) error {
	for key, cfg := range configs {
		s.Configs[key] = cfg
		if ct := constraint.NewConstraintsFrom(cfg); nil != ct {
			s.constraints[ct.ID] = ct
			// generate search indexes.
			searchIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagSearch)
			if len(searchIndexes) > 0 {
				s.searchConstraints =
					SliceAppend(s.searchConstraints, searchIndexes)
			}
			// generate time-series indexes.
			tseriesIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagTimeSeries)
			if len(tseriesIndexes) > 0 {
				s.tseriesConstraints =
					SliceAppend(s.tseriesConstraints, tseriesIndexes)
			}
		}
	}
	return nil
}

// RemoveConfigs remove entity property config.
func (s *statem) RemoveConfigs(propertyIDs []string) error {
	for _, propertyID := range propertyIDs {
		delete(s.Configs, propertyID)
		// TODO: 删除解析后的.
	}
	return nil
}

func (s *statem) SetMessageHandler(msgHandler MessageHandler) {
	s.msgHandler = msgHandler
}

// OnMessage recive statem input messages.
func (s *statem) OnMessage(msg Message) bool {
	var (
		reqID     = uuid()
		attaching = false
	)

	log.Debug("statem.OnMessage", logger.EntityID(s.ID), logger.RequestID(reqID))

	if s.status == SMStatusDeleted {
		return false
	}

	for {
		// 如果只有一条投递线程，那么会导致Dispatcher上的所有Entity都依赖于Message Queue中的消息的均匀性.
		if nil == s.mailBox.Put(msg) {
			break
		}

		runtime.Gosched()
	}

	if atomic.CompareAndSwapInt32(&s.attached, StateDetached, StateAttached) {
		attaching = true
		log.Info("attatched statem.", logger.EntityID(s.ID))
	}

	return attaching
}

// InvokeMsg run loopHandler.
func (s *statem) HandleLoop() { //nolint
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

		switch msg := message.(type) {
		case MapperMessage:
			s.invokeMapperMsg(msg)
		case TentacleMsg:
			s.invokeTentacleMsg(msg)
		default:
			// dispose message.
			watchKeys := s.msgHandler(message)
			// active tentacles.
			s.activeTentacle(watchKeys)
		}

		message.Promised(s)

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
					messages[targetID][active.PropertyKey] = thisStateProps[active.PropertyKey]
				} else {
					// undefined tentacle typs.
					log.Warn("undefined tentacle type", zap.Any("tentacle", tentacle))
				}
			}
		} else {
			// TODO...
			// 如果消息是缓存，那么，我们应该对改state的tentacles刷新。
			log.Debug("match end of \".*\" PropertyKey.", zap.String("entity", active.EntityId), zap.String("property-key", active.PropertyKey))
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
func (s *statem) activeMapper(actives map[string][]mapper.Tentacler) { //nolint
	if len(actives) == 0 {
		return
	}

	var err error
	for mapperID := range actives {
		input := make(map[string]constraint.Node)
		for _, tentacle := range s.indexTentacles[mapperID] {
			for _, item := range tentacle.Items() {
				var val constraint.Node
				if val, err = s.getProperty(s.cacheProps[item.EntityId], item.PropertyKey); nil != err {
					log.Error("get property failed", logger.RequestID(item.PropertyKey), zap.Error(err))
					continue
				} else if nil != val {
					input[item.String()] = val
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
					log.Error("get property failed",
						logger.EntityID(s.ID), zap.String("property_key", propertyKey), zap.Error(err))
				}
				s.LastTime = time.Now().UnixNano() / 1e6
			}
		}
	}
}

func (s *statem) getProperty(properties map[string]constraint.Node, propertyKey string) (constraint.Node, error) {
	if !strings.ContainsAny(propertyKey, ".[") {
		return s.KValues[propertyKey], nil
	}

	arr := strings.SplitN(propertyKey, ".", 2)
	// patch property.
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

// invokeMapperMsg dispose mapper msg.
func (s *statem) invokeMapperMsg(msg MapperMessage) {
	var err error
	switch msg.Operator {
	case MapperOperatorAppend:
		err = s.appendMapper(msg.Mapper)
	case MapperOperatorRemove:
		err = s.removeMapper(msg.Mapper)
	default:
		err = errInvalidMapperOp
	}

	if nil != err {
		log.Error("invoke mapper",
			zap.Error(err),
			logger.EntityID(s.ID),
			logger.TQLString(msg.Mapper.TQLString),
			logger.MapperID(s.ID+"#"+msg.Mapper.Name))
	} else {
		log.Debug("invoke mapper",
			logger.EntityID(s.ID),
			logger.TQLString(msg.Mapper.TQLString),
			logger.MapperID(s.ID+"#"+msg.Mapper.Name))
	}
	log.Info("recv mapper message", logger.EntityID(s.GetID()), zap.Any("mapper", msg.Mapper))
}

// SetMapper set mapper into entity.
func (s *statem) appendMapper(desc MapperDesc) error {
	reqID := uuid()

	// checked befors.
	m, _ := mapper.NewMapper(s.ID+"#"+desc.Name, desc.TQLString)

	log.Info("append mapper",
		logger.EntityID(s.ID), logger.RequestID(reqID), logger.MapperID(m.ID()), logger.TQLString(m.String()))

	position, length := 0, len(s.Mappers)
	for ; position < length; position++ {
		if desc.Name == s.Mappers[position].Name {
			s.Mappers[position].TQLString = desc.TQLString
			break
		}
	}

	if position < length {
		// 更新mapper之前我们需要将前面建立的删除
	} else {
		s.Mappers = append(s.Mappers, desc)
	}

	s.mappers[m.ID()] = m

	// generate indexTentacles again.
	for _, mp := range s.mappers {
		for _, tentacle := range mp.Tentacles() {
			s.indexTentacles[tentacle.TargetID()] =
				append(s.indexTentacles[tentacle.TargetID()], tentacle)
		}
	}

	// generate tentacles again.
	s.generateTentacles()

	sourceEntities := []string{}
	for _, expr := range m.SourceEntities() {
		sourceEntities = append(sourceEntities,
			s.stateManager.EscapedEntities(expr)...)
	}

	log.Info("source entities", zap.Any("entities", sourceEntities))
	for _, entityID := range sourceEntities {
		tentacle := mapper.MergeTentacles(s.indexTentacles[entityID]...)
		log.Info("record tentacle", logger.EntityID(s.ID), zap.String("target", entityID), zap.Any("tentacle", tentacle))
		if nil != tentacle {
			// send tentacle msg.
			s.stateManager.SendMsg(MessageContext{
				Headers: Header{
					MessageCtxHeaderSourceID: s.ID,
					MessageCtxHeaderTargetID: entityID,
				},
				Message: TentacleMsg{
					StateID:  s.ID,
					Operator: TentacleOperatorAppend,
					Items:    tentacle.Copy().Items(),
				},
			})
		}
	}

	return nil
}

func (s *statem) removeMapper(desc MapperDesc) error {
	position, length := 0, len(s.Mappers)
	for ; position < length; position++ {
		if desc.Name == s.Mappers[position].Name {
			break
		}
	}

	if position == length {
		return nil
	}

	m := s.mappers[s.ID+"#"+s.Mappers[position].Name]

	log.Info("remove mapper",
		logger.EntityID(s.ID), logger.MapperID(m.ID()), logger.TQLString(m.String()))

	// 这一块暂时这样做，但是实际上是存在问题的： tentacles创建和删除的顺序行，不同entity中tentacle的一致性问题，这个问题可以使用version来解决,此外如果tentacles是动态生成也会存在问题.
	// 如果是动态生成的，那么前后两次生成可能不一致.
	// 且这里使用了两个锁，存在死锁风险.
	sourceEntities := []string{m.TargetEntity()}
	for _, expr := range m.SourceEntities() {
		sourceEntities = append(sourceEntities,
			s.stateManager.EscapedEntities(expr)...)
	}

	for _, entityID := range sourceEntities {
		tentacle := mapper.MergeTentacles(s.indexTentacles[entityID]...)

		if nil != tentacle {
			// send tentacle msg.
			s.stateManager.SendMsg(MessageContext{
				Headers: Header{
					MessageCtxHeaderSourceID: s.ID,
					MessageCtxHeaderTargetID: entityID,
				},
				Message: &TentacleMsg{
					StateID:  s.ID,
					Operator: TentacleOperatorRemove,
					Items:    tentacle.Copy().Items(),
				},
			})
		}
	}
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
				log.Info("setup tentacle",
					logger.EntityID(s.ID), logger.Type(tentacle.Type()), logger.Target(tentacle.TargetID()))
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
