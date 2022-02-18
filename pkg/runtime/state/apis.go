package state

import (
	"context"
	"encoding/json"
	"sort"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/core/pkg/constraint"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type APIID string

func (a APIID) String() string {
	return string(a)
}

const (
	APICreateEntity  APIID = "corecreateentity"
	APIGetEntity     APIID = "coregetentity"
	APISetProperties APIID = "coresetproperties"
	APIPatchEntity   APIID = "corepatchentity"
	APIDeleteEntity  APIID = "coredeleteentity"
	APISetConfigs    APIID = "coresetconfigs"
	APIPatchConfigs  APIID = "corepatchconfigs"
)

type APIHandler func(context.Context, message.Context) []WatchKey

var once = sync.Once{}
var apiCallbacks map[APIID]APIHandler

// call core.APIs.
func (s *statem) callAPIs(ctx context.Context, msgCtx message.Context) []WatchKey {
	log.Debug("call core.APIs", zfield.Header(msgCtx.Attributes()))

	once.Do(func() {
		apiCallbacks = map[APIID]APIHandler{
			APICreateEntity: s.cbCreateEntity,
			APISetConfigs:   s.cbSetConfigs,
			APIPatchConfigs: s.cbPatchConfigs,
			APIDeleteEntity: s.cbDeleteEntity,
		}
	})

	apiID := APIID(msgCtx.Get(message.ExtAPIIdentify))
	return apiCallbacks[apiID](ctx, msgCtx)
}

func (s *statem) cbCreateEntity(ctx context.Context, msgCtx message.Context) []WatchKey {
	msgID := util.UUID()
	eventID := util.UUID()
	ev := cloudevents.NewEvent()

	ev.SetID(eventID)
	ev.SetType(s.Type)
	ev.SetSource(s.Source)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtEntityID, s.ID)
	ev.SetExtension(message.ExtEntityType, s.Type)
	ev.SetExtension(message.ExtEntityOwner, s.Owner)
	ev.SetExtension(message.ExtMessageReceiver, s.ID)
	ev.SetExtension(message.ExtEntitySource, s.Source)
	ev.SetExtension(message.ExtAPIRespStatus, types.StatusOK.String())
	ev.SetExtension(message.ExtCallback, msgCtx.Get(message.ExtCallback))
	ev.SetExtension(message.ExtAPIRequestID, msgCtx.Get(message.ExtAPIRequestID))
	ev.SetExtension(message.ExtMessageType, message.MessageTypeAPIRespond.String())
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	ev.SetData(msgCtx.Message())

	if err := ev.Validate(); nil != err {
		log.Error("validate response", zfield.Eid(s.ID), zap.Error(err))
		ev.SetExtension(message.ExtAPIRespStatus, types.StatusError.String())
		ev.SetExtension(message.ExtAPIRespErrCode, err.Error())
	}

	log.Debug("core.APIS callback", zfield.ID(s.ID), zfield.ReqID(msgCtx.Get(message.ExtAPIRequestID)))

	s.dispatcher.Dispatch(ctx, ev)

	return nil
}

func (s *statem) cbSetConfigs(ctx context.Context, msgCtx message.Context) []WatchKey {
	s.ConfigFile = msgCtx.Message()
	return nil
}

func (s *statem) cbPatchConfigs(ctx context.Context, msgCtx message.Context) []WatchKey { //nolint
	var (
		err       error
		bytes     []byte
		patchData []*PatchData
	)

	// TODO: decode message to PatchData.

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

func (s *statem) cbDeleteEntity(ctx context.Context, msgCtx message.Context) []WatchKey {
	s.status = SMStatusDeleted
	return nil
}
