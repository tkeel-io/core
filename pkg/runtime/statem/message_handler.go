package statem

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

// internelMessageHandler dispose statem input messages.
func (s *statem) internelMessageHandler(message Message) []WatchKey {
	switch msg := message.(type) {
	case PropertyMessage:
		return s.invokePropertyMessage(msg)
	case FlushPropertyMessage:
		// convert msg type.
		message := PropertyMessage(msg)
		// handle property message.
		watchKeys := s.internelMessageHandler(message)
		// flush state.
		s.flush(context.Background())
		return watchKeys
	default:
		// invalid msg typs.
		log.Error("undefine message type", zfield.ID(s.ID), zfield.Message(msg))
	}

	return nil
}

// invokePropertyMessage invoke property message.
func (s *statem) invokePropertyMessage(msg PropertyMessage) []WatchKey {
	setStateID := msg.StateID
	watchKeys := make([]mapper.WatchKey, 0)
	if _, has := s.cacheProps[setStateID]; !has {
		s.cacheProps[setStateID] = make(map[string]constraint.Node)
	}

	stateProps := s.cacheProps[setStateID]
	for key, value := range msg.Properties {
		if _, err := patchProperty(stateProps, key, constraint.PatchOpReplace, value); nil != err {
			log.Error("set state property", zfield.ID(s.ID), zfield.PK(key), zap.Error(err))
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
					MsgCtxHeaderSender:   s.ID,
					MsgCtxHeaderReceiver: stateID,
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
	for mapperID := range actives {
		input := make(map[string]constraint.Node)
		for _, tentacle := range s.mappers[mapperID].Tentacles() {
			for _, item := range tentacle.Items() {
				var val constraint.Node
				if val, err = s.getProperty(s.cacheProps[item.EntityId], item.PropertyKey); nil != err {
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

		var properties map[string]constraint.Node

		// excute mapper.
		if properties, err = s.mappers[mapperID].Exec(input); nil != err {
			log.Error("exec statem mapper failed ", zap.Error(err))
		}

		log.Debug("exec mapper", zfield.Mid(mapperID), zap.Any("input", input), zap.Any("output", properties))

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

func patchProperty(props map[string]constraint.Node, path string, op constraint.PatchOperator, val constraint.Node) (constraint.Node, error) {
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
