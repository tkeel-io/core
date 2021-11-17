package statem

// statem: state marchins.

import (
	"context"
	"crypto/rand"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
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
	ID       string                       `json:"id"`
	Type     string                       `json:"type"`
	Owner    string                       `json:"owner"`
	Status   string                       `json:"status"`
	Source   string                       `json:"source"`
	Version  int64                        `json:"version"`
	LastTime int64                        `json:"last_time"`
	Mappers  []MapperDesc                 `json:"mappers"`
	KValues  map[string]constraint.Node   `json:"properties"` //nolint
	Configs  map[string]constraint.Config `json:"configs"`
}

func (b *Base) Copy() Base {
	return Base{
		ID:       b.ID,
		Type:     b.Type,
		Owner:    b.Owner,
		Status:   b.Status,
		Source:   b.Source,
		Version:  b.Version,
		LastTime: b.LastTime,
		Mappers:  b.Mappers,
		KValues:  b.KValues,
	}
}

// statem state marchins.
type statem struct {
	Base

	// mapper & tentacles.
	mappers        map[string]mapper.Mapper              // key=mapperId
	tentacles      map[string][]mapper.Tentacler         // key=Sid#propertyKey
	cacheProps     map[string]map[string]constraint.Node // cache other property.
	indexTentacles map[string][]mapper.Tentacler         // key=targetId(mapperId/Sid)

	// mailbox & state runtime status.
	mailBox      *mailbox
	attached     int32
	disposing    int32
	stateManager StateManager
	msgHandler   MessageHandler

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
		Base: Base{
			ID:      in.ID,
			Type:    in.Type,
			Owner:   in.Owner,
			Status:  StateStatusActive,
			KValues: make(map[string]constraint.Node),
			Configs: make(map[string]constraint.Config),
		},

		ctx:            ctx,
		cancel:         cancel,
		stateManager:   stateMgr,
		msgHandler:     msgHandler,
		mailBox:        newMailbox(10),
		disposing:      StateDisposingIdle,
		mappers:        make(map[string]mapper.Mapper),
		cacheProps:     make(map[string]map[string]constraint.Node),
		indexTentacles: make(map[string][]mapper.Tentacler),
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

func (s *statem) GetManager() StateManager {
	return s.stateManager
}

func (s *statem) SetConfig(configs map[string]constraint.Config) error {
	for k, c := range configs {
		s.Configs[k] = c
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

	log.Infof("statem.OnMessage called, entityId: %s, requestId: %s, msg", s.ID, reqID)

	for {
		// 如果只有一条投递线程，那么会导致Dispatcher上的所有Entity都依赖于Message Queue中的消息的均匀性.
		if nil == s.mailBox.Put(msg) {
			break
		}

		runtime.Gosched()
	}

	if atomic.CompareAndSwapInt32(&s.attached, StateDetached, StateAttached) {
		attaching = true
		log.Infof("attatched statem, id: %s.", s.ID)
	}

	return attaching
}

// InvokeMsg run loopHandler.
func (s *statem) HandleLoop() {
	var (
		Ensure  = 3
		message Message
	)

	for {
		// consume message from mailbox.
		if message = s.mailBox.Get(); nil == message {
			if Ensure > 0 {
				Ensure--
				runtime.Gosched()
				continue
			}

			// detach this statem.
			if !atomic.CompareAndSwapInt32(&s.attached, StateAttached, StateDetached) {
				log.Errorf("exception occurred, mismatched statem runtime status: %s.",
					stateRuntimeStatusString(atomic.LoadInt32(&s.attached)))
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
			watchKeys := s.internelMessageHandler(message)
			// active tentacles.
			s.activeTentacle(watchKeys)
		}

		log.Infof("disposed message: %v", message)

		// reset be surs.
		Ensure = 3
	}

	log.Infof("detached statem, id: %s.", s.ID)
}

// InvokeMsg dispose statem input messages.
func (s *statem) internelMessageHandler(message Message) []WatchKey {
	switch msg := message.(type) {
	case PropertyMessage:
		return s.invokePropertyMsg(msg)
	default:
		// invalid msg typs.
		log.Errorf("undefine message type, msg: %s", msg)
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
		stateProps[key] = constraint.RawNode(value)
		watchKeys = append(watchKeys, mapper.WatchKey{EntityId: setStateID, PropertyKey: key})
	}

	// set last active tims.
	if setStateID == s.ID {
		s.LastTime = time.Now().UnixNano() / 1e6
	}

	return watchKeys
}

// activeTentacle active tentacles.
func (s *statem) activeTentacle(actives []mapper.WatchKey) {
	var (
		messages        = make(map[string]map[string][]byte)
		activeTentacles = make(map[string][]mapper.Tentacler)
	)

	thisStateProps := s.cacheProps[s.ID]
	for _, active := range actives {
		if tentacles, exists := s.tentacles[active.String()]; exists {
			for _, tentacle := range tentacles {
				targetID := tentacle.TargetID()
				if mapper.TentacleTypeMapper == tentacle.Type() {
					activeTentacles[targetID] = append(activeTentacles[targetID], tentacle)
				} else if mapper.TentacleTypeEntity == tentacle.Type() {
					// make if not exists.
					if _, exists := messages[targetID]; !exists {
						messages[targetID] = make(map[string][]byte)
					}

					// 在组装成Msg后，SendMsg的时候会对消息进行序列化，所以这里不需要Deep Copy.
					// 在这里我们需要解析PropertyKey, PropertyKey中可能存在嵌套层次.
					messages[targetID][active.PropertyKey] = []byte(thisStateProps[active.PropertyKey].String())
				} else {
					// undefined tentacle typs.
					log.Warnf("undefined tentacle type, %v", tentacle)
				}
			}
		} else { //nolint
			// TODO...
			// 如果消息是缓存，那么，我们应该对改state的tentacles刷新。
		}
	}

	for stateID, msg := range messages {
		s.stateManager.SendMsg(MessageContext{
			Headers: Header{
				MessageCtxHeaderSourceID: s.ID,
				MessageCtxHeaderTargetID: stateID,
			},
			Message: &PropertyMessage{
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
	for mapperID := range actives {
		input := make(map[string]constraint.Node)
		for _, tentacle := range s.indexTentacles[mapperID] {
			for _, item := range tentacle.Items() {
				input[item.String()] = s.getProperty(s.cacheProps[item.EntityId], item.PropertyKey)
			}
		}

		// excute mapper.
		properties, err := s.mappers[mapperID].Exec(input)
		if nil != err {
			log.Errorf("exec statem mapper failed ", err)
		}

		log.Infof("exec mapper: %s, output: %v", mapperID, properties)

		if len(properties) > 0 {
			for propertyKey, value := range properties {
				s.setProperty(propertyKey, value)
			}
			s.LastTime = time.Now().UnixNano() / 1e6
		}
	}
}

func (s *statem) getProperty(properties map[string]constraint.Node, propertyKey string) constraint.Node {
	if len(properties) == 0 {
		return nil
	}

	// 我们或许应该在这里解析propertyKey中的嵌套层次.
	return properties[propertyKey]
}

func (s *statem) setProperty(propertyKey string, value constraint.Node) {
	// 我们或许应该在这里解析propertyKey中的嵌套层次.
	if _, has := s.KValues[propertyKey]; !has {
		s.KValues[propertyKey] = value
		return
	}
	s.KValues[propertyKey] = value
}

// invokeMapperMsg dispose mapper msg.
func (s *statem) invokeMapperMsg(msg MapperMessage) {
	var err error
	switch msg.Operator {
	case MapperOperatorAppend:
		s.appendMapper(msg.Mapper)
	case MapperOperatorRemove:
		s.removeMapper(msg.Mapper)
	default:
		err = errInvalidMapperOp
	}

	log.Infof("invoke mapper, mapperId: %s, stateId: %s, tql: %s, error: %v",
		s.ID+"#"+msg.Mapper.Name, s.ID, msg.Mapper.TQLString, err)
}

// SetMapper set mapper into entity.
func (s *statem) appendMapper(desc MapperDesc) error {
	reqID := uuid()

	// checked befors.
	m, _ := mapper.NewMapper(s.ID+"#"+desc.Name, desc.TQLString)

	log.Infof("append mapper, stateID: %s, requestId: %s, mapperId: %s, mapper: %s.",
		s.ID, reqID, m.ID, m.String())

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
		log.Errorf("invalid tentacle operator: %s, %v", msg.Operator, msg)
	}
	log.Infof("catch tentacle event, op: %s, target: %s, msg: %v.", msg.Operator, msg.StateID, msg)

	// generate tentacles again.
	s.generateTentacles()
}

func (s *statem) generateTentacles() {
	s.tentacles = make(map[string][]mapper.Tentacler)
	for _, tentacles := range s.indexTentacles {
		for _, tentacle := range tentacles {
			if mapper.TentacleTypeMapper == tentacle.Type() || tentacle.IsRemote() {
				log.Infof("%s set up tentacle, type: %s, target: %s.", s.ID, tentacle.Type(), tentacle.TargetID())
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
