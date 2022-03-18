package state

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/collectjs/pkg/json/jsonparser"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/util"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

func (s *statem) getState(stateID string) State {
	if s.ID == stateID {
		return State{ID: s.ID, Props: s.Properties}
	}

	if _, ok := s.cacheProps[stateID]; !ok {
		s.cacheProps[stateID] = make(map[string]tdtl.Node)
	}

	return State{ID: stateID, Props: s.cacheProps[stateID]}
}

type tsDevice struct {
	TS     int64              `json:"ts,omitempty"`
	Values map[string]float64 `json:"values,omitempty"`
}

type tsData struct {
	TS    int64   `json:"ts,omitempty"`
	Value float64 `json:"value,omitempty"`
}

func adjustTSData(bytes []byte) (dataAdjust []byte) {
	// tsDevice1 no ts
	tsDevice1 := make(map[string]float64)
	err := json.Unmarshal(bytes, &tsDevice1)
	if err == nil && len(tsDevice1) > 0 {
		tsDeviceAdjustData := make(map[string]*tsData)
		for k, v := range tsDevice1 {
			tsDeviceAdjustData[k] = &tsData{TS: time.Now().UnixMilli(), Value: v}
		}
		dataAdjust, _ = json.Marshal(tsDeviceAdjustData)
		log.Info(string(dataAdjust))
		return
	}

	// tsDevice2 has ts
	tsDevice2 := tsDevice{}
	err = json.Unmarshal(bytes, &tsDevice2)
	log.Info(tsDevice2)
	if err == nil && tsDevice2.TS != 0 {
		tsDeviceAdjustData := make(map[string]*tsData)
		for k, v := range tsDevice2.Values {
			tsDeviceAdjustData[k] = &tsData{TS: tsDevice2.TS, Value: v}
		}
		dataAdjust, _ = json.Marshal(tsDeviceAdjustData)
		log.Info(string(dataAdjust))
		return
	}

	tsGatewayData := make(map[string]*tsDevice)
	//		tsGatewayAdjustData := make(map[string]map[string]*tsData)
	tsGatewayAdjustData := make(map[string]interface{})

	err = json.Unmarshal(bytes, &tsGatewayData)
	if err == nil {
		for k, v := range tsGatewayData {
			tsGatewayAdjustDataK := map[string]*tsData{}
			for kk, vv := range v.Values {
				tsGatewayAdjustDataK[kk] = &tsData{TS: v.TS, Value: vv}
				//		tsGatewayAdjustData[strings.Join([]string{k, kk}, ".")] = &tsData{TS: v.TS, Value: vv}
			}
			tsGatewayAdjustData[k] = tsGatewayAdjustDataK
		}
		dataAdjust, _ = json.Marshal(tsGatewayAdjustData)
		log.Info(string(dataAdjust))
		return
	}
	log.Error("ts data adjust error", zap.Error(err))
	return dataAdjust
}

func (s *statem) invokeRawMessage(ctx context.Context, msgCtx message.Context) []WatchKey {
	log.Debug("invoke raw message", zfield.Eid(s.ID), zfield.Type(s.Type),
		zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())))

	var err error
	var rawData RawData
	// decode raw data json.
	if err = json.Unmarshal(msgCtx.Message(), &rawData); nil != err {
		log.Error("decode raw data", zfield.Eid(s.ID), zfield.Type(s.Type), zap.Error(err),
			zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())))
		return nil
	}

	actives := make([]WatchKey, 0)
	actives = append(actives, WatchKey{EntityID: s.ID, PropertyKey: "rawData"})
	stateIns := State{ID: s.ID, Props: s.Properties}

	// set raw data.
	stateIns.Patch(xjson.OpReplace, "rawData", msgCtx.Message())
	if rawData.Type == "rawData" {
		return actives
	}

	// dispose rawdata.
	var bytes []byte
	if bytes, err = base64.StdEncoding.DecodeString(rawData.Values); nil != err {
		log.Error("decode raw data", zfield.Eid(s.ID), zfield.Type(s.Type), zap.Error(err),
			zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())))
		return nil
	}

	if rawData.Type == "telemetry" {
		dataAdjust := adjustTSData(bytes)
		bytes = dataAdjust

		/*
			oldTelemetry, err := stateIns.Get("telemetry")
			if err == nil {
				dataMerged, err = collectjs.Merge(oldTelemetry.Raw(), dataAdjust)
				if err != nil {
					log.Error("ts data merge error", zap.Error(err))

				} else {
					dataMerged = dataAdjust
				}
			}
		*/
	}

	collect := collectjs.ByteNew(bytes)
	if err = collect.GetError(); nil != err {
		log.Warn("raw data content type unknown", zfield.Eid(s.ID), zfield.Type(s.Type),
			zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())), zfield.Reason(err.Error()))
		return nil
	} else if jsonparser.Object.String() != collect.GetDataType() {
		log.Warn("raw data content type unknown", zfield.Eid(s.ID), zfield.Type(s.Type),
			zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())))
		return nil
	}

	collect.Foreach(func(key, value []byte, dataType jsonparser.ValueType) {
		path := strings.Join([]string{rawData.Type, strings.Trim(string(key), `"`)}, ".")
		if _, err = stateIns.Patch(xjson.OpReplace, path, value); nil != err {
			log.Error("decode raw data", zfield.Eid(s.ID), zfield.Type(s.Type), zap.Error(err),
				zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())))
			return
		}
		actives = append(actives, WatchKey{EntityID: s.ID, PropertyKey: path})
	})

	return actives
}

// invokePropertyMessage invoke property message.
func (s *statem) invokeStateMessage(ctx context.Context, msgCtx message.Context) []WatchKey {
	stateID := msgCtx.Get(message.ExtEntityID)
	log.Debug("invoke state message", zfield.Eid(s.ID), zfield.Type(s.Type),
		zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())))

	stateIns := s.getState(stateID)
	watchKeys := make([]mapper.WatchKey, 0)
	collectjs.ForEach(msgCtx.Message(), jsonparser.Object,
		func(key, value []byte, dataType jsonparser.ValueType) {
			propertyKey := string(key)
			if _, err := stateIns.Patch(xjson.OpReplace, propertyKey, value); nil != err {
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

func (s *statem) invokeRepublishMessage(ctx context.Context, msgCtx message.Context) []WatchKey {
	stateID := msgCtx.Get(message.ExtEntityID)
	msgSender := msgCtx.Get(message.ExtSenderID)
	log.Debug("invoke republish message", zfield.Eid(s.ID), zfield.Type(s.Type),
		zfield.Header(msgCtx.Attributes()), zfield.Message(string(msgCtx.Message())))

	stateIns := s.getState(msgSender)
	watchKeys := make([]mapper.WatchKey, 0)
	collectjs.ForEach(msgCtx.Message(), jsonparser.Object,
		func(key, value []byte, dataType jsonparser.ValueType) {
			propertyKey := string(key)
			if _, err := stateIns.Patch(xjson.OpReplace, propertyKey, value); nil != err {
				log.Error("upsert state property", zfield.ID(s.ID), zfield.PK(propertyKey), zap.Error(err))
			} else {
				watchKeys = append(watchKeys, mapper.WatchKey{EntityID: msgSender, PropertyKey: propertyKey})
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

	log.Debug("active state tentacle", zfield.Eid(s.ID),
		zfield.Type(s.Type), zap.Any("actives", actives), zap.Any("tetacles", s.tentacles))

	var (
		messages        = make(map[string]map[string]tdtl.Node)
		activeTentacles = make([]string, 0)
	)

	for _, active := range actives {
		// full match.
		stateIns := s.getState(active.EntityID)
		if tentacles, exists := s.tentacles[active.String()]; exists {
			for _, tentacle := range tentacles {
				targetID := tentacle.TargetID()
				if mapper.TentacleTypeMapper == tentacle.Type() {
					activeTentacles = append(activeTentacles, tentacle.TargetID())
				} else if mapper.TentacleTypeEntity == tentacle.Type() {
					// make if not exists.
					if _, exists := messages[targetID]; !exists {
						messages[targetID] = make(map[string]tdtl.Node)
					}

					if prop, err := stateIns.Patch(xjson.OpCopy, active.PropertyKey, nil); nil != err {
						log.Warn("patch copy property", zfield.Eid(s.ID), zfield.Value(stateIns.Props),
							zfield.Target(targetID), zfield.PK(active.String()), zfield.Reason(err.Error()))
					} else if nil != prop {
						messages[targetID][active.PropertyKey] = prop
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
							activeTentacles = append(activeTentacles, targetID)
						} else if mapper.TentacleTypeEntity == tentacle.Type() {
							// make if not exists.
							if _, exists := messages[targetID]; !exists {
								messages[targetID] = make(map[string]tdtl.Node)
							}

							stateIns := s.getState(active.EntityID)
							segments := strings.Split(active.PropertyKey, ".")
							if prop, has := stateIns.Props[segments[0]]; has {
								messages[targetID][segments[0]] = prop
							} else {
								log.Warn("patch copy property", zfield.Eid(s.ID), zfield.Value(stateIns.Props),
									zfield.Target(targetID), zfield.PK(active.String()), zfield.Reason("path not found"))
							}
						}
					}
				}
			}
		}
	}

	for stateID, msg := range messages {
		ev := cloudevents.NewEvent()
		ev.SetID(util.UUID("ev"))
		ev.SetType("republish")
		ev.SetSource("core.runtime")
		ev.SetExtension(message.ExtEntityID, stateID)
		ev.SetExtension(message.ExtEntityType, s.Type)
		ev.SetExtension(message.ExtEntityOwner, s.Owner)
		ev.SetExtension(message.ExtEntitySource, s.Source)
		ev.SetExtension(message.ExtSenderID, s.ID)
		ev.SetExtension(message.ExtMessageReceiver, stateID)
		ev.SetDataContentType(cloudevents.ApplicationJSON)

		var err error
		var bytes []byte
		// encode message.

		msgArr := []string{}
		for key, val := range msg {
			msgArr = append(msgArr, fmt.Sprintf("\"%s\":%s", key, string(val.Raw())))
		}

		bytes = []byte(fmt.Sprintf("{%s}", strings.Join(msgArr, ",")))

		if err = ev.SetData(bytes); nil != err {
			log.Error("set event payload", zap.Error(err), zfield.Eid(s.ID))
			continue
		} else if err = ev.Validate(); nil != err {
			log.Error("validate event", zap.Error(err), zfield.Eid(s.ID))
			continue
		}

		log.Debug("republish message", zap.String("event_id", ev.Context.GetID()), zfield.Event(ev))

		s.dispatcher.Dispatch(context.Background(), ev)
	}

	// active mapper.
	s.activeMapper(activeTentacles)
}

// activeMapper active mappers.
func (s *statem) activeMapper(actives []string) {
	if len(actives) == 0 {
		return
	}

	log.Debug("active state mapper", zfield.Eid(s.ID),
		zfield.Type(s.Type), zap.Strings("actives", actives))

	// unique slice.
	actives = util.Unique(actives)

	var err error
	var activeKeys []mapper.WatchKey
	for _, mapperID := range actives {
		input := make(map[string]tdtl.Node)
		for _, tentacle := range s.mappers[mapperID].Tentacles() {
			for _, item := range tentacle.Items() {
				var val tdtl.Node
				stateIns := s.getState(item.EntityID)
				if val, err = stateIns.Patch(xjson.OpCopy, item.PropertyKey, nil); nil != err {
					log.Warn("patch copy", zfield.Reason(err.Error()), zfield.Mid(mapperID), zap.String("dispose_entity", s.ID),
						zfield.Eid(item.EntityID), zfield.PK(item.PropertyKey), zfield.Value(stateIns.Props))
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

		stateIns := s.getState(s.ID)
		for propertyKey, value := range properties {
			setVal := []byte(string(value.Raw()))
			if _, err = stateIns.Patch(xjson.OpReplace, propertyKey, setVal); nil != err {
				log.Error("set property", zfield.ID(s.ID), zap.Error(err),
					zap.String("property_key", propertyKey), zap.String("value", string(setVal)))
				continue
			}
			s.LastTime = time.Now().UnixNano() / 1e6
			activeKeys = append(activeKeys, mapper.WatchKey{EntityID: s.ID, PropertyKey: propertyKey})
		}
	}

	// TODO: 这里需要注意循环调用.
	s.activeTentacle(unique(activeKeys))
}

func (s *statem) invokeMapperInit(ctx context.Context, msgCtx message.Context) []WatchKey {
	var (
		err      error
		actives  []string
		messages = make(map[string]map[string]tdtl.Node)
	)

	for _, tentacle := range s.sCtx.tentacles {
		// inie mapper tentacle.
		if tentacle.Version() == 0 {
			targetID := tentacle.TargetID()
			switch tentacle.Type() {
			case mapper.TentacleTypeMapper:
				actives = append(actives, tentacle.TargetID())
			case mapper.TentacleTypeEntity:
				for _, item := range tentacle.Items() {
					var res tdtl.Node
					stateIns := s.getState(item.EntityID)
					if res, err = stateIns.Get(item.PropertyKey); nil != err {
						log.Warn("init tentacle, patch copy",
							zap.String("dispose_entity", item.EntityID),
							zfield.Eid(s.ID), zfield.PK(item.String()),
							zfield.Reason(err.Error()), zap.Any("cache", s.cacheProps),
							zfield.Value(stateIns.Props), zfield.Path(item.String()))
						continue
					}

					if _, ok := messages[targetID]; !ok {
						messages[targetID] = make(map[string]tdtl.Node)
					}

					// set message.
					messages[targetID][item.PropertyKey] = res
				}
			}
		}
	}

	// republish messages.
	for stateID, msg := range messages {
		ev := cloudevents.NewEvent()
		ev.SetID(util.UUID("ev"))
		ev.SetType("republish")
		ev.SetSource("core.runtime")
		ev.SetExtension(message.ExtEntityID, stateID)
		ev.SetExtension(message.ExtSenderID, s.ID)
		ev.SetExtension(message.ExtSenderType, s.Type)
		ev.SetExtension(message.ExtSenderOwner, s.Owner)
		ev.SetExtension(message.ExtSenderSource, s.Source)
		ev.SetExtension(message.ExtMessageReceiver, stateID)
		ev.SetDataContentType(cloudevents.ApplicationJSON)

		var err error
		var bytes []byte
		// encode message.

		msgArr := []string{}
		for key, val := range msg {
			msgArr = append(msgArr, fmt.Sprintf("\"%s\":%s", key, string(val.Raw())))
		}

		bytes = []byte(fmt.Sprintf("{%s}", strings.Join(msgArr, ",")))

		if err = ev.SetData(bytes); nil != err {
			log.Error("set event payload", zap.Error(err), zfield.Eid(s.ID))
			continue
		} else if err = ev.Validate(); nil != err {
			log.Error("validate event", zap.Error(err), zfield.Eid(s.ID))
			continue
		}

		s.dispatcher.Dispatch(context.Background(), ev)
	}

	s.activeMapper(actives)

	return nil
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
