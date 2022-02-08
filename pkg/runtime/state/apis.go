package state

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/core/pkg/constraint"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

// call core.APIs.
func (s *statem) callAPIs(ctx context.Context, msgCtx message.Context) error {
	log.Debug("call core.APIs", zfield.Header(msgCtx.Attributes()))

	switch msg := msgCtx.Message().(type) {
	case message.StateMessage:
		// handle state machine message.
		switch msg.Method {
		case message.SMMethodSetConfigs:
			return errors.Wrap(s.callSetConfigs(ctx, msgCtx), "call core.APIs.SetConfigs")
		case message.SMMethodPatchConfigs:
			return errors.Wrap(s.callPatchConfigs(ctx, msgCtx), "call core.APIs.PatchConfigs")
		case message.SMMethodDeleteEntity:
			s.status = SMStatusDeleted
		default:
			log.Error("core.APIs not support",
				zfield.Method(msg.Method.String()))
		}
	default:
		log.Error("invalid state message type")
	}

	return nil
}

func (s *statem) callSetConfigs(ctx context.Context, msgCtx message.Context) error {
	var err error
	var bytes []byte
	stateMessage, _ := msgCtx.Message().(message.StateMessage)
	if bytes, err = json.Marshal(stateMessage.Value); nil != err {
		log.Error("json marshal", zap.Error(err), zfield.Header(msgCtx.Attributes()))
		return errors.Wrap(err, "json marshal")
	}

	s.ConfigFile = bytes
	return nil
}

func (s *statem) callPatchConfigs(ctx context.Context, msgCtx message.Context) error { //nolint
	var (
		err       error
		bytes     []byte
		patchData []*PatchData
	)

	// assert state Message.
	stateMessage, _ := msgCtx.Message().(message.StateMessage)
	if err = mapstructure.Decode(stateMessage.Value, &patchData); nil != err {
		log.Error("decode patch data", zap.Error(err), zfield.Header(msgCtx.Attributes()))
		return errors.Wrap(err, "call core.APIs.PatchConfigs")
	}

	// marshal config.
	for _, pd := range patchData {
		if pd.Value, err = json.Marshal(pd.Value); nil != err {
			log.Error("json marshal", zap.Error(err), zfield.Header(msgCtx.Attributes()))
		}
	}

	for _, pd := range patchData {
		switch pd.Operator {
		case constraint.PatchOpAdd:
			if bytes, err = collectjs.Append(s.ConfigFile, pd.Path, pd.Value.([]byte)); nil != err {
				log.Error("call core.APIs.PatchConfigs patch add", zap.Error(err))
				continue
			}
			s.ConfigFile = bytes
		case constraint.PatchOpRemove:
			s.ConfigFile = collectjs.Del(s.ConfigFile, pd.Path)
		case constraint.PatchOpReplace:
			if bytes, err = collectjs.Set(s.ConfigFile, pd.Path, pd.Value.([]byte)); nil != err {
				log.Error("call core.APIs.PatchConfigs patch add", zap.Error(err))
				continue
			}
			s.ConfigFile = bytes
		}
	}

	// parse state config again.
	configs := make(map[string]interface{})
	if err = json.Unmarshal(s.ConfigFile, &configs); nil != err {
		log.Error("json unmarshal", zap.Error(err), zap.String("configs", string(s.ConfigFile)))
		return errors.Wrap(err, "json unmarshal")
	}

	var cfg constraint.Config
	cfgs := make(map[string]constraint.Config)
	for key, val := range configs {
		if cfg, err = constraint.ParseConfigFrom(val); nil != err {
			log.Error("parse configs", zap.Error(err))
			continue
		}
		cfgs[key] = cfg
	}

	// reset state machine configs.
	s.constraints = make(map[string]*constraint.Constraint)
	s.searchConstraints = make(sort.StringSlice, 0)
	s.tseriesConstraints = make(sort.StringSlice, 0)

	for _, cfg = range cfgs {
		if ct := constraint.NewConstraintsFrom(cfg); nil != ct {
			s.constraints[ct.ID] = ct
			// generate search indexes.
			if searchIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagSearch); len(searchIndexes) > 0 {
				s.searchConstraints = util.SliceAppend(s.searchConstraints, searchIndexes)
			}

			// generate time-series indexes.
			if tseriesIndexes := ct.GenEnabledIndexes(constraint.EnabledFlagTimeSeries); len(tseriesIndexes) > 0 {
				s.tseriesConstraints = util.SliceAppend(s.tseriesConstraints, tseriesIndexes)
			}
		}
	}

	return nil
}
