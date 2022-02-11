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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/resource/search"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
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
	var flushData = make(map[string]interface{})
	for _, JSONPath := range s.searchConstraints {
		var val constraint.Node
		var ct *constraint.Constraint
		if val, err = s.getProperty(s.Properties, JSONPath); nil != err {
			log.Warn("flush search", zap.Error(err), zfield.ID(s.ID))
		} else if ct, err = s.getConstraint(JSONPath); nil != err {
			// TODO: 终止本次写入.
		} else if val, err = constraint.ExecData(val, ct); nil != err {
			// TODO: 终止本次写入.
		} else if nil == val {
			flushData[JSONPath] = val.Value()
			continue
		}
		log.Warn("patch.copy entity property failed",
			zfield.Eid(s.ID), zap.String("property_key", JSONPath), zap.Error(err))
	}

	// flush all.
	for key, val := range s.Properties {
		flushData[key] = val.Value()
	}

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

func (s *statem) flushTimeSeries(ctx context.Context) error {
	var err error
	var flushData []tseries.TSeriesData
	for _, JSONPath := range s.tseriesConstraints {
		var val constraint.Node
		var ct *constraint.Constraint
		if val, err = s.getProperty(s.Properties, JSONPath); nil != err || val == nil {
		} else if ct, err = s.getConstraint(JSONPath); nil != err {
		} else if val, err = constraint.ExecData(val, ct); nil != err || val == nil {
		} else {
			point := tseries.TSeriesData{
				Measurement: "core-default",
				Tags:        s.generateTags(),
				Fields:      map[string]string{},
				Value:       val.String(),
			}
			flushData = append(flushData, point)
		}
		log.Warn("patch.copy entity property failed", zfield.Eid(s.ID), zap.String("property_key", JSONPath), zap.Error(err))
	}

	if _, err = s.TSeries().Write(ctx, convertPoints(flushData)); nil != err {
		log.Error("flush timeseries Search.", zap.Any("data", flushData), zap.Error(err))
	}

	log.Debug("flush timeseries Search.", zap.Any("data", flushData))
	return errors.Wrap(err, "timeseries flush failed")
}

// generateTags generate entity tags.
func (s *statem) generateTags() map[string]string {
	return map[string]string{
		"app":    "core",
		"id":     s.ID,
		"type":   s.Type,
		"owner":  s.Owner,
		"source": s.Source,
	}
}

func (s *statem) getConstraint(jsonPath string) (*constraint.Constraint, error) {
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
	return s.stateManager.Resource().TSDB()
}

func (s *statem) Pubsub() pubsub.Pubsub {
	return s.stateManager.Resource().Pubsub()
}

func (s *statem) Search() *search.Service {
	return s.stateManager.Resource().Search()
}

func (s *statem) Repo() repository.IRepository {
	return s.stateManager.Resource().Repo()
}

func convertPoints(points []tseries.TSeriesData) *tseries.TSeriesRequest {
	lines := make([]string, 0)
	for _, point := range points {
		point.Fields["value"] = point.Value
		lines = append(lines,
			fmt.Sprintf("%s,%s %s", point.Measurement, util.ExtractMap(point.Tags), util.ExtractMap(point.Fields)))
	}
	return &tseries.TSeriesRequest{Data: lines}
}
