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

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *statem) Flush(ctx context.Context) error {
	return s.flush(ctx)
}

func (s *statem) FlushState() error {
	return errors.Wrap(s.flushState(s.ctx), "flush state-machine state")
}

func (s *statem) FlushSearch() error {
	return errors.Wrap(s.flushSearch(s.ctx), "flush state-machine state")
}

func (s *statem) FlushTimeSeries() error {
	return errors.Wrap(s.flushTimeSeries(s.ctx), "flush state-machine time-series")
}

func (s *statem) flush(ctx context.Context) error {
	var err error
	// flush state properties to es.
	if err = s.flushSearch(ctx); nil == err {
		log.Debug("entity flush Search completed", zfield.Eid(s.ID))
	}
	// flush state properties to state.
	if err = s.flushState(ctx); nil == err {
		log.Debug("entity flush State completed", zfield.Eid(s.ID))
	}
	// flush state properties to TSDB.
	if err = s.flushTimeSeries(ctx); nil == err {
		log.Debug("entity flush TimeSeries completed", zfield.Eid(s.ID))
	}
	return errors.Wrap(err, "entity flush data failed")
}

func (s *statem) flushState(ctx context.Context) error {
	log.Debug("flush state",
		zfield.Eid(s.ID),
		zfield.Type(s.Type),
		zfield.Template(s.TemplateID),
		zap.String("state", s.Entity.JSON()))

	var err error
	if err = s.Repo().PutEntity(ctx, &s.Entity); nil == err {
		log.Debug("entity flush state", zap.Error(err), zfield.Eid(s.ID))
	}

	return errors.Wrap(err, "flush entity state")
}

func (s *statem) flushSearch(ctx context.Context) error {
	var err error
	// flush all.
	flushData := s.toGolang(s.Properties)

	// basic fields.
	flushData["id"] = s.ID
	flushData["type"] = s.Type
	flushData["owner"] = s.Owner
	flushData["source"] = s.Source
	flushData["version"] = s.Version
	flushData["last_time"] = s.LastTime

	var data *structpb.Value
	if data, err = structpb.NewValue(flushData); nil != err {
		log.Error("flush state Search.", zap.Any("data", flushData), zap.Error(err))
		return errors.Wrap(err, "Search flush")
	} else if _, err = s.Search().Index(ctx, &pb.IndexObject{Obj: data}); nil != err {
		log.Error("flush state Search.", zap.Any("data", flushData), zap.Error(err))
	}

	log.Debug("flush state Search.", zap.Any("data", flushData))
	return errors.Wrap(err, "Search flush")
}

func (s *statem) toGolang(vals map[string]tdtl.Node) map[string]interface{} {
	result := make(map[string]interface{})
	for key, val := range vals {
		var res interface{}
		switch val.Type() {
		case tdtl.Undefined:
		case tdtl.Null:
		case tdtl.Bool:
			res = (val.String() == "true")
		case tdtl.Number:
		case tdtl.Int:
			i, err := strconv.ParseInt(val.String(), 10, 63)
			if nil != err {
				log.Warn("parse integer type", zap.Error(err))
			}
			res = i
		case tdtl.Float:
			i, err := strconv.ParseFloat(val.String(), 64)
			if nil != err {
				log.Warn("parse float type", zap.Error(err))
			}
			res = i
		case tdtl.String:
			res = val.String()
		case tdtl.JSON, tdtl.Array, tdtl.Object:
			err := json.Unmarshal(val.Raw(), &res)
			if nil != err {
				log.Warn("parse json type", zap.Error(err))
			}
		default:
			log.Warn("type error", val.Type())
		}
		result[key] = res
	}
	return result
}

func (s *statem) flushTimeSeries(ctx context.Context) error {
	var err error
	sData := s.toGolang(s.Properties)
	log.Info("sData: ", sData)
	tsData, ok := sData["telemetry"]
	if !ok {
		return nil
	}
	var flushData []*tseries.TSeriesData
	log.Info("tsData: ", tsData)
	tt, ok := tsData.(map[string]interface{})
	if ok {
		for k, v := range tt {
			switch ttt := v.(type) {
			case map[string]interface{}:
				if ts, ok := ttt["ts"]; ok {
					tsItem := tseries.TSeriesData{
						Measurement: "keel",
						Tags:        map[string]string{"id": s.ID},
						Fields:      map[string]float32{},
						Timestamp:   0,
					}
					switch tttV := ttt["value"].(type) {
					case float64:
						tsItem.Fields[k] = float32(tttV)
						tsItem.Timestamp = int64(ts.(float64)) * 1e6
						flushData = append(flushData, &tsItem)
					case float32:
						tsItem.Fields[k] = tttV
						tsItem.Timestamp = int64(ts.(float64)) * 1e6
						flushData = append(flushData, &tsItem)
					}
					continue
				}
			default:
				log.Info(ttt)
			}
		}
	}
	s.TSeries().Write(ctx, &tseries.TSeriesRequest{
		Data:     flushData,
		Metadata: map[string]string{},
	})
	return errors.Wrap(err, "timeseries flush failed")
}

// generateTags generate entity tags.
func (s *statem) generateTags() map[string]string { // no lint
	return map[string]string{
		"app":    "core",
		"id":     s.ID,
		"type":   s.Type,
		"owner":  s.Owner,
		"source": s.Source,
	}
}

func (s *statem) getConstraint(jsonPath string) (*constraint.Constraint, error) { // no lint
	arr := strings.Split(jsonPath, ".")
	if len(arr) == 0 {
		return nil, xerrors.ErrInvalidJSONPath
	} else if len(arr) == 1 {
		return s.constraints[arr[0]], nil
	}

	var ct *constraint.Constraint
	if ct = s.constraints[arr[0]]; nil != ct {
		return nil, nil
	}

	var index int
	for indx, key := range arr[1:] {
		var nextCt *constraint.Constraint
		for _, childCt := range ct.ChildNodes {
			if key == childCt.ID {
				nextCt, index = childCt, indx+1
				break
			}
		}
		if nextCt == nil {
			break
		}
		ct = nextCt
	}

	if index != len(arr)-1 {
		return nil, nil
	}

	return ct, nil
}

func (s *statem) TSeries() tseries.TimeSerier {
	return s.resourceManager.TSDB()
}

func (s *statem) Search() *search.Service {
	return s.resourceManager.Search()
}

func (s *statem) Repo() repository.IRepository {
	return s.resourceManager.Repo()
}
