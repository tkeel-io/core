package state

import (
	"context"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	"github.com/tkeel-io/core/pkg/constraint"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

func (s *statem) getState(stateID string) State {
	if _, ok := s.cacheProps[stateID]; !ok {
		s.cacheProps[stateID] = make(map[string]tdtl.Node)
	}

	return State{ID: stateID, Props: s.cacheProps[stateID]}
}

// invokePropertyMessage invoke property message.
func (s *statem) invokePropertyMessage(msgCtx message.Context) []WatchKey {
	stateID := msgCtx.Get(message.ExtEntityID)
	if _, has := s.cacheProps[stateID]; !has {
		s.cacheProps[stateID] = make(map[string]tdtl.Node)
	}

	stateIns := s.getState(stateID)
	watchKeys := make([]mapper.WatchKey, 0)
	collectjs.ForEach(msgCtx.Message(), jsonparser.Object, func(key, value []byte) {
		propertyKey := string(key)
		if _, err := stateIns.Patch(constraint.OpReplace, propertyKey, value); nil != err {
			log.Error("upsert state property", zfield.ID(s.ID), zfield.PK(propertyKey), zap.Error(err))
		} else {
			watchKeys = append(watchKeys, mapper.WatchKey{EntityID: stateID, PropertyKey: propertyKey})
		}
	})

	// set last active tims.
	if stateID == s.ID {
		s.Version++
		s.LastTime = util.UnixMilli()
	}

	return watchKeys
}

// activeTentacle active tentacles.
func (s *statem) activeTentacle(actives []mapper.WatchKey) { //nolint
	if len(actives) == 0 {
		return
	}

	var (
		messages        = make(map[string]map[string]tdtl.Node)
		activeTentacles = make(map[string][]mapper.Tentacler)
	)

	for _, active := range actives {
		// full match.
		stateIns := s.getState(active.EntityID)
		if tentacles, exists := s.tentacles[active.String()]; exists {
			for _, tentacle := range tentacles {
				targetID := tentacle.TargetID()
				if mapper.TentacleTypeMapper == tentacle.Type() {
					activeTentacles[targetID] = append(activeTentacles[targetID], tentacle)
				} else if mapper.TentacleTypeEntity == tentacle.Type() {
					// make if not exists.
					if _, exists := messages[targetID]; !exists {
						messages[targetID] = make(map[string]tdtl.Node)
					}

					// 在组装成Msg后，SendMsg的时候会对消息进行序列化，所以这里不需要Deep Copy.
					// 在这里我们需要解析PropertyKey, PropertyKey中可能存在嵌套层次.

					if prop, err := stateIns.Patch(constraint.OpCopy, active.PropertyKey, nil); nil == err {
						messages[targetID][active.PropertyKey] = prop
					} else {
						log.Warn("patch copy property", zfield.Eid(s.ID), zfield.PK(active.PropertyKey))
					}
				} else {
					// undefined tentacle typs.
					log.Warn("undefined tentacle type", zap.Any("tentacle", tentacle))
				}
			}
		} else {
			// TODO: topic 规则匹配树.
			// 如果消息是缓存，那么，我们应该对改state的tentacles刷新。
			log.Debug("match end of string \".*\" PropertyKey.", zap.String("entity", active.EntityID), zap.String("property-key", active.PropertyKey))
			// match entityID.*   .
			for watchKey, tentacles := range s.tentacles {
				arr := strings.Split(watchKey, ".")
				if len(arr) == 2 && arr[1] == "*" && arr[0] == active.EntityID {
					for _, tentacle := range tentacles {
						targetID := tentacle.TargetID()
						if mapper.TentacleTypeMapper == tentacle.Type() {
							activeTentacles[targetID] = append(activeTentacles[targetID], tentacle)
						} else if mapper.TentacleTypeEntity == tentacle.Type() {
							// make if not exists.
							if _, exists := messages[targetID]; !exists {
								messages[targetID] = make(map[string]tdtl.Node)
							}

							segments := strings.Split(active.PropertyKey, ".")
							// 在组装成Msg后，SendMsg的时候会对消息进行序列化，所以这里不需要Deep Copy.
							// 在这里我们需要解析PropertyKey, PropertyKey中可能存在嵌套层次.
							// TODO:
							messages[targetID][segments[0]] = s.getState(active.EntityID).Props[segments[0]]
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
		ev := cloudevents.NewEvent()
		ev.SetID(util.UUID())
		ev.SetType("republish")
		ev.SetSource("core.runtime")
		ev.SetExtension(message.ExtMessageSender, s.ID)
		ev.SetExtension(message.ExtMessageReceiver, stateID)
		ev.SetDataContentType(cloudevents.ApplicationJSON)

		var err error
		var bytes []byte
		// encode message.

		if bytes, err = constraint.EncodeJSON(msg); nil != err {
			log.Error("encode state properties", zap.Error(err), zfield.Eid(s.ID))
		}

		ev.SetData(bytes)
		log.Debug("republish message", zap.String("event_id", ev.Context.GetID()), zfield.Event(ev))

		s.dispatcher.Dispatch(context.Background(), ev)
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
	var stateIns = s.getState(s.ID)
	var activeKeys []mapper.WatchKey
	for mapperID := range actives {
		input := make(map[string]tdtl.Node)
		for _, tentacle := range s.mappers[mapperID].Tentacles() {
			for _, item := range tentacle.Items() {
				var val tdtl.Node
				if val, err = stateIns.Patch(constraint.OpCopy, item.PropertyKey, nil); nil != err {
					log.Error("patch copy", zfield.ReqID(item.PropertyKey), zap.Error(err))
					continue
				} else if nil != val {
					input[item.String()] = val
				}
			}
		}

		if len(input) == 0 {
			log.Debug("obtain mapper input, empty params", zfield.Mid(mapperID))
			continue
		}

		var properties map[string]tdtl.Node

		// excute mapper.
		if properties, err = s.mappers[mapperID].Exec(input); nil != err {
			log.Error("exec statem mapper failed ", zap.Error(err))
		}

		log.Debug("exec mapper", zfield.Mid(mapperID), zap.Any("input", input), zap.Any("output", properties))

		for propertyKey, value := range properties {
			if _, err = stateIns.Patch(constraint.OpReplace, propertyKey, []byte(value.String())); nil != err {
				log.Error("set property", zfield.ID(s.ID), zap.String("property_key", propertyKey), zap.Error(err))
				continue
			}
			s.LastTime = time.Now().UnixNano() / 1e6
			activeKeys = append(activeKeys, mapper.WatchKey{EntityID: s.ID, PropertyKey: propertyKey})
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
