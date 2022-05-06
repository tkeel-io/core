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

package runtime

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/tdtl"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/kit/log"
)

func (n *Node) FlushEntity(ctx context.Context, en Entity) error {
	log.L().Debug("flush entity", logf.Eid(en.ID()), logf.Value(string(en.Raw())))

	// 1. flush state.
	if err := n.resourceManager.Repo().PutEntity(ctx, en.ID(), en.Raw()); nil != err {
		log.L().Error("flush entity state storage", logf.Error(err), logf.Eid(en.ID()))
		return errors.Wrap(err, "flush entity into state storage")
	}

	// 2. flush data.
	// 2.1 flush search global data.
	globalData, err := n.makeSearchData(en)
	if nil != err {
		log.L().Error("make SearchData error", logf.Error(err), logf.Eid(en.ID()))
	} else {
		if _, err = n.resourceManager.Search().IndexBytes(ctx, en.ID(), globalData); nil != err {
			log.L().Error("flush entity search engine", logf.Error(err), logf.Eid(en.ID()))
			//			return errors.Wrap(err, "flush entity into search engine")
		}
	}

	// 2.2 flush search model data.
	// TODO.

	// 2.3 flush timeseries data.
	flushData, err := n.makeTimeSeriesData(ctx, en)
	if nil != err {
		log.L().Error("make TimeSeries error", logf.Error(err), logf.Eid(en.ID()))
	} else {
		if _, err = n.resourceManager.TSDB().Write(ctx, flushData); nil != err {
			log.L().Error("flush entity timeseries database", logf.Error(err), logf.Eid(en.ID()))
			//			return errors.Wrap(err, "flush entity into search engine")
		}
	}

	// 2.4 flush raw data.
	rawData, err := n.makeRawData(ctx, en)
	if nil != err {
		log.L().Error("make RawData error", logf.Error(err), logf.Eid(en.ID()))
	} else {
		if err := n.resourceManager.RawData().Write(context.Background(), rawData); nil != err {
			log.L().Error("flush entity rawData", logf.Error(err), logf.Eid(en.ID()))
		}
	}

	return nil
}

func (n *Node) makeRawData(ctx context.Context, en Entity) (*rawdata.Request, error) {
	req := &rawdata.Request{}
	req.Metadata = make(map[string]string)
	raw := en.GetProp("rawData")

	req.Metadata["path"] = en.GetProp("rawData.path").String()
	req.Metadata["type"] = en.GetProp("rawData.type").String()
	req.Metadata["mark"] = en.GetProp("rawData.mark").String()

	tsStr := en.GetProp("rawData.ts").String()
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return nil, err
	}
	req.Data = append(req.Data, &rawdata.RawData{
		EntityID:  en.ID(),
		Path:      "rawData",
		Values:    string(raw.Raw()),
		Timestamp: time.UnixMilli(ts),
	})
	return req, nil
}

func (n *Node) makeTimeSeriesData(ctx context.Context, en Entity) (*tseries.TSeriesRequest, error) {
	tsData := en.GetProp("telemetry")
	var flushData []*tseries.TSeriesData
	log.Info("tsData: ", tsData)
	var res interface{}

	err := json.Unmarshal(tsData.Raw(), &res)
	if nil != err {
		log.L().Warn("parse json type", logf.Error(err))
		return nil, errors.Wrap(err, "write ts db error")
	}
	tss, ok := res.(map[string]interface{})
	if ok {
		for k, v := range tss {
			switch tsOne := v.(type) {
			case map[string]interface{}:
				if ts, ok := tsOne["ts"]; ok {
					tsItem := tseries.TSeriesData{
						Measurement: "keel",
						Tags:        map[string]string{"id": en.ID()},
						Fields:      map[string]float32{},
						Timestamp:   0,
					}
					switch tttV := tsOne["value"].(type) {
					case float64:
						tsItem.Fields[k] = float32(tttV)
						timestamp, _ := ts.(float64)
						tsItem.Timestamp = int64(timestamp) * 1e6
						flushData = append(flushData, &tsItem)
					case float32:
						tsItem.Fields[k] = tttV
						timestamp, _ := ts.(float64)
						tsItem.Timestamp = int64(timestamp) * 1e6
						flushData = append(flushData, &tsItem)
					}
					continue
				}
			default:
				log.Info(tsOne)
			}
		}
	}
	return &tseries.TSeriesRequest{
		Data:     flushData,
		Metadata: map[string]string{},
	}, errors.Wrap(err, "write ts db error")
}

func (n *Node) makeSearchData(en Entity) ([]byte, error) {
	globalData := collectjs.ByteNew([]byte(`{}`))
	globalData.Set(FieldID, en.Get(FieldID).Raw())
	globalData.Set(FieldType, en.Get(FieldType).Raw())
	globalData.Set(FieldOwner, en.Get(FieldOwner).Raw())
	globalData.Set(FieldSource, en.Get(FieldSource).Raw())
	globalData.Set(FieldTemplate, en.Get(FieldTemplate).Raw())

	byt, err := json.Marshal(string(en.Raw()))
	if err != nil {
		log.L().Error("json marshal error")
	}
	globalData.Set(FieldEntitySource, byt)

	sysField := en.GetProp("sysField")
	if sysField.Type() != tdtl.Null {
		globalData.Set("sysField", sysField.Raw())
	}
	basicInfo := en.GetProp("basicInfo")
	if basicInfo.Type() != tdtl.Null {
		globalData.Set("basicInfo", basicInfo.Raw())
	}
	connectInfo := en.GetProp("connectInfo")
	if connectInfo.Type() != tdtl.Null {
		globalData.Set("connectInfo", connectInfo.Raw())
	}
	return globalData.GetRaw(), nil
}

func (n *Node) RemoveEntity(ctx context.Context, en Entity) error {
	var err error

	// recover entity state.
	defer func() {
		if nil != err {
			if innerErr := n.FlushEntity(ctx, en); nil != innerErr {
				log.L().Error("remove entity failed, recover entity state failed", logf.Eid(en.ID()),
					logf.Reason(err.Error()), logf.Error(innerErr), logf.Value(string(en.Raw())))
			}
		}
	}()

	// 1. 从状态存储中删除（可标记）
	if err := n.resourceManager.Repo().
		DelEntity(ctx, en.ID()); nil != err {
		log.L().Error("remove entity from state storage",
			logf.Error(err), logf.Eid(en.ID()), logf.Value(string(en.Raw())))
		return errors.Wrap(err, "remove entity from state storage")
	}

	// 2. 从搜索中删除（可标记）
	if _, err := n.resourceManager.Search().
		DeleteByID(ctx, &v1.DeleteByIDRequest{
			Id:     en.ID(),
			Owner:  en.Owner(),
			Source: en.Source(),
		}); nil != err {
		log.L().Error("remove entity from state search engine",
			logf.Error(err), logf.Eid(en.ID()), logf.Value(string(en.Raw())))
		return errors.Wrap(err, "remove entity from state search engine")
	}

	// 3. 删除实体相关的 Expression.
	return nil
}
