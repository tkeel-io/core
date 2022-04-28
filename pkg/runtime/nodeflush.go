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

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func (n *Node) FlushEntity(ctx context.Context, en Entity) error {
	log.L().Debug("flush entity", zfield.Eid(en.ID()), zfield.Value(string(en.Raw())))

	// 1. flush state.
	if err := n.resourceManager.Repo().PutEntity(ctx, en.ID(), en.Raw()); nil != err {
		log.L().Error("flush entity state storage", zap.Error(err), zfield.Eid(en.ID()))
		return errors.Wrap(err, "flush entity into state storage")
	}

	// 2. flush search engine data.
	// 2.1 flush search global data.
	globalData := n.getGlobalData(en)
	if _, err := n.resourceManager.Search().IndexBytes(ctx, en.ID(), globalData); nil != err {
		log.L().Error("flush entity search engine", zap.Error(err), zfield.Eid(en.ID()))
		//			return errors.Wrap(err, "flush entity into search engine")
	}

	// 2.2 flush search model data.
	// TODO.

	// 3. flush timeseries data.
	if err := n.flushTimeSeries(ctx, en); nil != err {
		log.L().Error("flush entity timeseries database", zap.Error(err), zfield.Eid(en.ID()))
	}

	// 4. flush raw data.
	if err := n.flushRawData(ctx, en); nil != err {
		log.L().Error("flush entity rawData", zap.Error(err), zfield.Eid(en.ID()))
	}
	return nil
}

func (n *Node) flushRawData(ctx context.Context, en Entity) (err error) {
	req := &rawdata.RawDataRequest{}
	req.Metadata = make(map[string]string)
	raw := en.GetProp("rawData")

	req.Metadata["path"] = en.GetProp("rawData.path").String()
	req.Metadata["type"] = en.GetProp("rawData.type").String()
	req.Metadata["mark"] = en.GetProp("rawData.mark").String()

	tsStr := en.GetProp("rawData.ts").String()
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, &rawdata.RawData{
		EntityID:  en.ID(),
		Path:      "rawData",
		Values:    string(raw.Raw()),
		Timestamp: time.UnixMilli(ts),
	})
	n.resourceManager.RawData().Write(context.Background(), req)
	return nil
}

func (n *Node) flushTimeSeries(ctx context.Context, en Entity) (err error) {
	tsData := en.GetProp("telemetry")
	var flushData []*tseries.TSeriesData
	log.Info("tsData: ", tsData)
	var res interface{}

	err = json.Unmarshal(tsData.Raw(), &res)
	if nil != err {
		log.L().Warn("parse json type", zap.Error(err))
		return
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
	_, err = n.resourceManager.TSDB().Write(ctx, &tseries.TSeriesRequest{
		Data:     flushData,
		Metadata: map[string]string{},
	})
	return errors.Wrap(err, "write ts db error")
}

func (n *Node) RemoveEntity(ctx context.Context, en Entity) error {
	var err error

	// recover entity state.
	defer func() {
		if nil != err {
			if innerErr := n.FlushEntity(ctx, en); nil != innerErr {
				log.L().Error("remove entity failed, recover entity state failed", zfield.Eid(en.ID()),
					zfield.Reason(err.Error()), zap.Error(innerErr), zfield.Value(string(en.Raw())))
			}
		}
	}()

	// 1. 从状态存储中删除（可标记）
	if err := n.resourceManager.Repo().
		DelEntity(ctx, en.ID()); nil != err {
		log.L().Error("remove entity from state storage",
			zap.Error(err), zfield.Eid(en.ID()), zfield.Value(string(en.Raw())))
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
			zap.Error(err), zfield.Eid(en.ID()), zfield.Value(string(en.Raw())))
		return errors.Wrap(err, "remove entity from state search engine")
	}

	// 3. 删除实体相关的 Expression.
	return nil
}
