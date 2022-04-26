package runtime

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/mapper/expression"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/core/pkg/util/path"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

const (
	rawDataRawType       = "rawData"
	rawDataTelemetryType = "telemetry"
)

type EntityResourceFunc func(context.Context, Entity) error

type EntityResource struct {
	FlushHandler  EntityResourceFunc
	RemoveHandler EntityResourceFunc
}

type Runtime struct {
	id              string
	evalTree        *path.Tree
	subTree         *path.RefTree
	enCache         EntityCache
	entities        map[string]Entity // 存放Runtime的实体.
	dispatcher      dispatch.Dispatcher
	expressions     map[string]ExpressionInfo
	repository      repository.IRepository
	entityResourcer EntityResource

	mlock  sync.RWMutex
	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRuntime(ctx context.Context, ercFuncs EntityResource, id string, dispatcher dispatch.Dispatcher, repository repository.IRepository) *Runtime {
	ctx, cancel := context.WithCancel(ctx)
	return &Runtime{
		id:              id,
		enCache:         NewCache(repository),
		entities:        map[string]Entity{},
		expressions:     map[string]ExpressionInfo{},
		entityResourcer: ercFuncs,
		dispatcher:      dispatcher,
		repository:      repository,
		subTree:         path.NewRefTree(),
		evalTree:        path.New(),
		lock:            sync.RWMutex{},
		mlock:           sync.RWMutex{},
		cancel:          cancel,
		ctx:             ctx,
	}
}

func (r *Runtime) ID() string {
	return r.id
}

func (r *Runtime) DeliveredEvent(ctx context.Context, msg *sarama.ConsumerMessage) {
	var err error
	var ev v1.ProtoEvent
	if err = v1.Unmarshal(msg.Value, &ev); nil != err {
		log.L().Error("decode Event", zap.Error(err),
			zfield.Message(string(msg.Value)), zfield.RID(r.id))
		return
	}

	r.HandleEvent(ctx, &ev)
}

func (r *Runtime) HandleEvent(ctx context.Context, event v1.Event) error {
	log.L().Debug("handle event", zfield.RID(r.id),
		zfield.Event(event), zfield.EvID(event.ID()))

	execer, feed := r.PrepareEvent(ctx, event)
	feed = execer.Exec(ctx, feed)

	// call callback once.
	r.handleCallback(ctx, feed)
	if nil != feed.Err {
		log.Error("handle event", zap.Error(feed.Err),
			zfield.ID(event.ID()), zfield.Eid(event.Entity()), zfield.Event(event))
	}

	return nil
}

func (r *Runtime) PrepareEvent(ctx context.Context, ev v1.Event) (*Execer, *Feed) {
	log.L().Info("prepare event", zfield.RID(r.id),
		zfield.ID(ev.ID()), zfield.Eid(ev.Entity()))

	switch ev.Type() {
	case v1.ETSystem:
		execer, feed := r.prepareSystemEvent(ctx, ev)
		execer.postFuncs = append(execer.postFuncs,
			&handlerImpl{fn: r.handlePersistent})
		return execer, feed
	case v1.ETEntity:
		e, _ := ev.(v1.PatchEvent)
		state, err := r.LoadEntity(ev.Entity())
		if nil != err {
			log.L().Error("load entity", zfield.Eid(ev.Entity()),
				zap.Error(err), zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
			state = DefaultEntity(ev.Entity())
		}

		execer := &Execer{
			state: state,
			preFuncs: []Handler{
				&handlerImpl{fn: r.handleRawData}},
			execFunc: state,
			postFuncs: []Handler{
				&handlerImpl{fn: r.handleTentacle},
				&handlerImpl{fn: r.handleComputed},
				&handlerImpl{fn: r.handlePersistent},
				&handlerImpl{fn: r.handleTemplate}}}

		return execer, &Feed{
			Err:      err,
			Event:    ev,
			State:    state.Raw(),
			EntityID: ev.Entity(),
			Patches:  conv(e.Patches())}
	case v1.ETCache:
		sender := ev.Attr(v1.MetaSender)
		// load cache.
		state, err := r.enCache.Load(ctx, sender)
		if nil != err {
			log.L().Error("load cache entity", zfield.Header(ev.Attributes()),
				zfield.Eid(ev.Entity()), zfield.ID(ev.ID()), zfield.Sender(sender))
			state = DefaultEntity(sender)
		}

		e, _ := ev.(v1.PatchEvent)
		execer := &Execer{
			state:     state,
			execFunc:  state,
			preFuncs:  []Handler{&handlerImpl{fn: r.handleSubscribe}},
			postFuncs: []Handler{&handlerImpl{fn: r.handleComputed}}}
		return execer, &Feed{
			Err:      err,
			Event:    ev,
			State:    state.Raw(),
			EntityID: sender,
			Patches:  conv(e.Patches())}
	default:
		return &Execer{}, &Feed{
			Event:    ev,
			State:    DefaultEntity("").Raw(),
			Err:      fmt.Errorf(" unknown RuntimeEvent Type"),
			EntityID: ev.Entity(),
		}
	}
}

func (r *Runtime) loadTemplate(tid string) (tdtl.Node, error) {
	if strings.TrimSpace(tid) == "" {
		return tdtl.New(`{}`), nil
	}

	ten, err := r.LoadEntity(tid)
	if nil != err {
		log.L().Error("load template", zap.Error(err), zfield.Eid(tid))
		return nil, errors.Wrap(err, "load template")
	}

	return ten.Get(FieldScheme), nil
}

// 处理实体生命周期.
func (r *Runtime) prepareSystemEvent(ctx context.Context, event v1.Event) (*Execer, *Feed) {
	log.L().Info("prepare system event", zfield.ID(event.ID()), zfield.Header(event.Attributes()))
	ev, _ := event.(v1.SystemEvent)
	action := ev.Action()
	operator := action.Operator
	switch v1.SystemOp(operator) {
	case v1.OpCreate:
		log.L().Info("create entity", zfield.Eid(ev.Entity()),
			zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))

		execer := &Execer{
			state:    DefaultEntity(ev.Entity()),
			preFuncs: []Handler{},
			execFunc: DefaultEntity(ev.Entity()),
			postFuncs: []Handler{
				&handlerImpl{fn: r.handleTentacle},
				&handlerImpl{fn: r.handleComputed},
				&handlerImpl{fn: func(_ context.Context, feed *Feed) *Feed {
					log.L().Info("create entity successed", zfield.Eid(ev.Entity()),
						zfield.ID(ev.ID()), zfield.Header(ev.Attributes()), zfield.Value(string(action.Data)))
					return feed
				}},
			}}

		// check entity exists.
		if _, exists := r.entities[ev.Entity()]; exists {
			return execer, &Feed{
				Event:    ev,
				EntityID: ev.Entity(),
				Err:      xerrors.ErrEntityAleadyExists}
		}

		// new entity.
		state, err := NewEntity(ev.Entity(), action.GetData())
		if nil != err {
			log.L().Error("create entity", zfield.Eid(ev.Entity()),
				zfield.Value(string(action.GetData())), zap.Error(err))
			return execer, &Feed{
				Err:      err,
				Event:    ev,
				EntityID: ev.Entity()}
		}

		var scheme tdtl.Node
		templateID := state.Get(FieldTemplate).String()
		if scheme, err = r.loadTemplate(templateID); nil != err {
			log.L().Error("load template", zap.Error(err),
				zfield.Eid(ev.Entity()), zfield.Template(templateID))
			return execer, &Feed{
				Err:      err,
				Event:    ev,
				EntityID: ev.Entity()}
		}

		props := state.Get(FieldProperties)
		r.entities[ev.Entity()] = state
		execer.state = state
		execer.execFunc = state
		return execer, &Feed{
			Err:      props.Error(),
			Event:    ev,
			State:    state.Raw(),
			EntityID: ev.Entity(),
			Patches: []Patch{{
				Op:    xjson.OpMerge,
				Path:  FieldProperties,
				Value: tdtl.New(props.Raw())}, {
				Op:    xjson.OpReplace,
				Path:  FieldScheme,
				Value: tdtl.New(scheme.Raw()),
			}}}
	case v1.OpDelete:
		state, err := r.LoadEntity(ev.Entity())
		if nil != err {
			state = DefaultEntity(ev.Entity())
			if errors.Is(err, xerrors.ErrEntityNotFound) {
				// TODO: if entity not exists.
				return &Execer{
						state:    state,
						execFunc: state,
					}, &Feed{
						Event:    ev,
						State:    state.Raw(),
						EntityID: ev.Entity()}
			}
			log.L().Error("delete entity", zfield.Eid(ev.Entity()),
				zfield.Value(string(action.GetData())), zap.Error(err))
		}

		execer := &Execer{
			state:    state,
			execFunc: state,
			preFuncs: []Handler{
				&handlerImpl{fn: func(ctx context.Context, feed *Feed) *Feed {
					if innerErr := r.entityResourcer.RemoveHandler(ctx, state); nil != innerErr {
						log.L().Error("delete entity failure", zfield.Eid(ev.Entity()),
							zap.Error(innerErr), zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
						feed.Err = innerErr
						return feed
					}

					// remove entity from runtime.
					delete(r.entities, state.ID())

					return feed
				}}},
			postFuncs: []Handler{
				&handlerImpl{fn: r.handleTentacle},
				&handlerImpl{fn: func(_ context.Context, feed *Feed) *Feed {
					log.L().Info("delete entity successed", zfield.Eid(ev.Entity()),
						zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
					return feed
				}}},
		}

		return execer, &Feed{
			Err:      err,
			Event:    ev,
			State:    state.Raw(),
			EntityID: ev.Entity()}
	default:
		return &Execer{
				state:    DefaultEntity(ev.Entity()),
				preFuncs: []Handler{},
				execFunc: DefaultEntity(ev.Entity()),
				postFuncs: []Handler{
					&handlerImpl{fn: func(_ context.Context, feed *Feed) *Feed {
						log.L().Error("event type not support", zfield.Eid(ev.Entity()),
							zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
						return feed
					}}}}, &Feed{
				Event:    ev,
				EntityID: ev.Entity(),
				Err:      xerrors.ErrInternal}
	}
}

func (r *Runtime) handleComputed(ctx context.Context, feed *Feed) *Feed {
	log.L().Debug("handle computed", zfield.Eid(feed.EntityID))
	// 1. 检查 ret.path 和 订阅列表.
	entityID := feed.EntityID
	expressions := make(map[string]ExpressionInfo)
	for _, change := range feed.Changes {
		for _, node := range r.evalTree.
			MatchPrefix(path.FmtWatchKey(entityID, change.Path)) {
			evalEnd, _ := node.(*EvalEndpoint)
			if expr, has := r.getExpr(evalEnd.expresionID); has {
				expressions[expr.ID] = expr
			}
		}
	}

	patches := make(map[string][]*v1.PatchData)
	for id, expr := range expressions {
		target := expr.EntityID

		// TODO: 当收到 ETEntity 类型的事件，事件不应该触发不属于自己的 Expression.
		if target != feed.EntityID && v1.ETEntity == feed.Event.Type() {
			continue
		}

		log.L().Debug("eval expression",
			zfield.Eid(entityID), zfield.Mid(id),
			zfield.Expr(expr.Expression.Expression))
		result, err := r.evalExpression(ctx, expr.Expression)
		if nil != err {
			log.L().Error("eval expression",
				zfield.Eid(entityID), zfield.Mid(id),
				zfield.Expr(expr.Expression.Expression))
			continue
		} else if nil == result {
			log.L().Warn("eval expression, empty result.",
				zfield.Eid(entityID), zfield.Mid(id),
				zfield.Expr(expr.Expression.Expression))
			continue
		}

		patches[target] = append(
			patches[target],
			&v1.PatchData{
				Operator: xjson.OpReplace.String(),
				Path:     expr.Expression.Path,
				Value:    result.Raw(),
			})
	}

	// 2. dispatch.send()
	for target, patch := range patches {
		r.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
			Id:        util.IG().EvID(),
			Timestamp: time.Now().UnixNano(),
			Metadata: map[string]string{
				v1.MetaType:     string(v1.ETEntity),
				v1.MetaBorn:     "handleComputed",
				v1.MetaEntityID: target},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: patch,
				}},
		})
	}

	return feed
}

func (r *Runtime) evalExpression(ctx context.Context, expr repository.Expression) (tdtl.Node, error) {
	var (
		err      error
		has      bool
		exprInfo ExpressionInfo
	)

	if exprInfo, has = r.getExpr(expr.ID); !has {
		log.L().Error("expression not exsists", zfield.ID(expr.ID),
			zfield.Eid(expr.EntityID), zfield.Expr(expr.Expression))
		return nil, xerrors.ErrExpressionNotFound
	}

	in := make(map[string]tdtl.Node)
	for _, item := range exprInfo.evalEndpoints {
		// entityID.propertyKey
		segs := strings.SplitN(item.path, ".", 2)

		var state Entity
		// get value from entities.
		if state, has = r.entities[segs[0]]; has {
			in[item.path] = state.Get(segs[1])
			continue
		}
		// get value from cache.
		if state, err = r.enCache.Load(ctx, segs[0]); nil == err {
			in[item.path] = state.Get(segs[1])
		}
	}

	// ignore empty input.
	if len(in) == 0 {
		log.L().Warn("ignore empty input",
			zfield.ID(expr.ID), zfield.Expr(expr.Expression))
		return nil, nil
	}

	exprIns, err := expression.NewExpr(exprInfo.Expression.Expression, nil)
	if nil != err {
		log.L().Error("parse expression",
			zfield.Eid(expr.EntityID), zap.Error(err))
		return nil, errors.Wrap(err, "parse expression")
	}

	// eval expression.
	var out tdtl.Node
	if out, err = exprIns.Eval(ctx, in); nil != err {
		log.L().Error("eval expression", zfield.Input(in),
			zfield.ID(expr.ID), zfield.Eid(expr.EntityID), zfield.Output(out.String()))
		return nil, errors.Wrap(err, "eval expression")
	}

	log.L().Debug("eval expression", zfield.ID(expr.ID),
		zfield.Eid(expr.EntityID), zfield.Input(in), zfield.Output(out))

	// clean nil feed.
	if out.Type() == tdtl.Null || out.Type() == tdtl.Undefined {
		log.L().Warn("invalid eval result", zfield.Eid(expr.EntityID),
			zap.Any("value", out.String()), zfield.ID(expr.ID), zfield.Expr(expr.Expression))
		return nil, xerrors.ErrInvalidParam
	}

	return out, nil
}

func mergePath(subPath, changePath string) string {
	// subPath format: entity_id.property_key
	seg2 := strings.SplitN(subPath, ".", 2)
	return path.MergePath(seg2[1], changePath)
}

func whichPrefix(targetPath, changePath string) string {
	if targetPath == "" || len(targetPath) > len(changePath) {
		return changePath
	}
	return targetPath
}

func (r *Runtime) handleTentacle(ctx context.Context, feed *Feed) *Feed {
	log.L().Debug("handle tentacle", zfield.Eid(feed.EntityID),
		zap.Any("changes", feed.Changes), zap.String("state", string(feed.State)))

	// 1. 检查 ret.path 和 订阅列表.
	targets := make(map[string]string)
	entityID := feed.EntityID
	var patches = make(map[string]*v1.PatchData)
	for _, change := range feed.Changes {
		for _, node := range r.subTree.
			MatchPrefix(path.FmtWatchKey(entityID, change.Path)) {
			subEnd, _ := node.(*SubEndpoint)
			subPath := mergePath(subEnd.path, change.Path)
			targets[subEnd.deliveryID] = whichPrefix(targets[subEnd.deliveryID], subPath)
			log.L().Debug("expression sub matched", zfield.Eid(entityID), zfield.Path(change.Path),
				zfield.Target(subEnd.target), zfield.Path(subEnd.path), zfield.ID(subEnd.deliveryID), zfield.Expr(subEnd.Expression()))
		}

		// TODO: 提到for外存在优化空间.
		for runtimeID, sendPath := range targets {
			if sendPath == change.Path {
				patches[runtimeID] = &v1.PatchData{
					Path:     change.Path,
					Operator: xjson.OpReplace.String(),
					Value:    change.Value.Raw(),
				}
				continue
			}

			// select send data.
			stateIns, _ := NewEntity(feed.EntityID, feed.State)
			sendVal := stateIns.Get(sendPath)
			if tdtl.Undefined != sendVal.Type() {
				patches[runtimeID] = &v1.PatchData{
					Path:     sendPath,
					Operator: xjson.OpReplace.String(),
					Value:    sendVal.Raw(),
				}
			}
		}
	}

	// 2. dispatch.send()
	for runtimeID, sendData := range patches {
		eventID := util.IG().EvID()
		log.L().Debug("republish event", zfield.ID(r.id), zfield.RID(r.id),
			zfield.EvID(eventID), zfield.Target(runtimeID), zfield.Value(sendData))

		// dispatch cache event.
		r.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
			Id:        eventID,
			Timestamp: time.Now().UnixNano(),
			Metadata: map[string]string{
				v1.MetaType:        string(v1.ETCache),
				v1.MetaBorn:        "handleTentacle",
				v1.MetaPartitionID: runtimeID,
				v1.MetaSender:      entityID},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: []*v1.PatchData{sendData},
				}},
		})
	}

	return feed
}

func (r *Runtime) handleCallback(ctx context.Context, feed *Feed) error {
	var err error
	event := feed.Event
	log.L().Debug("handle event callback.", zfield.ID(event.ID()),
		zfield.Eid(event.Entity()), zfield.Header(event.Attributes()))

	if event.CallbackAddr() != "" {
		if feed.Err == nil {
			// 需要注意的是：为了精炼逻辑，runtime内部只是对api返回变更后实体的最新状态，而不做API结果的组装.
			ev := &v1.ProtoEvent{
				Id:        event.ID(),
				Timestamp: time.Now().UnixNano(),
				Callback:  event.CallbackAddr(),
				Metadata:  event.Attributes(),
				Data: &v1.ProtoEvent_RawData{
					RawData: feed.State}}
			ev.SetType(v1.ETCallback)
			ev.SetAttr(v1.MetaBorn, "handleCallback")
			ev.SetAttr(v1.MetaResponseStatus, string(types.StatusOK))
			err = r.dispatcher.Dispatch(ctx, ev)
		} else {
			ev := &v1.ProtoEvent{
				Id:        event.ID(),
				Timestamp: time.Now().UnixNano(),
				Callback:  event.CallbackAddr(),
				Metadata:  event.Attributes(),
				Data: &v1.ProtoEvent_RawData{
					RawData: []byte(`{}`)}}
			ev.SetType(v1.ETCallback)
			ev.SetAttr(v1.MetaBorn, "handleCallback")
			ev.SetAttr(v1.MetaResponseErrCode, feed.Err.Error())
			ev.SetAttr(v1.MetaResponseStatus, string(types.StatusError))
			err = r.dispatcher.Dispatch(ctx, ev)
		}
	}

	if nil != err {
		log.L().Error("handle event, callback.", zfield.ID(event.ID()),
			zap.Error(err), zfield.Eid(event.Entity()), zfield.Header(event.Attributes()))
	}

	return errors.Wrap(err, "handle callback")
}

func (r *Runtime) handlePersistent(ctx context.Context, feed *Feed) *Feed {
	log.L().Debug("handle persistent", zfield.Eid(feed.EntityID))
	en, ok := r.entities[feed.EntityID]
	if !ok {
		// entity has been deleted.
		return feed
	}
	r.entityResourcer.FlushHandler(ctx, en)
	return feed
}

func (r *Runtime) handleTemplate(ctx context.Context, feed *Feed) *Feed {
	log.L().Debug("handle template", zfield.Eid(feed.EntityID))
	for index := range feed.Changes {
		if FieldTemplate == feed.Changes[index].Path {
			log.Info("entity template changed", zfield.Eid(feed.EntityID),
				zfield.Template(feed.Changes[index].Value.String()))
			feed.Err = r.onTemplateChanged(ctx,
				feed.EntityID, feed.Changes[index].Value.String())
			break
		}
	}
	return feed
}

func (r *Runtime) onTemplateChanged(ctx context.Context, entityID, templateID string) error {
	log.L().Debug("entity template changed", zfield.Eid(entityID), zfield.Template(templateID))
	// load template entity.
	templateIns, err := r.LoadEntity(templateID)
	if nil != err {
		log.L().Error("onTemplateChanged", zap.Error(err),
			zfield.Eid(entityID), zfield.Template(templateID))
		return errors.Wrap(err, "On Template Changed")
	}

	ev := &v1.ProtoEvent{
		Id:        entityID,
		Timestamp: time.Now().UnixNano(),
		Metadata: map[string]string{
			v1.MetaType:     string(v1.ETEntity),
			v1.MetaBorn:     "onTemplateChanged",
			v1.MetaEntityID: entityID,
			v1.MetaSender:   entityID},
		Data: &v1.ProtoEvent_Patches{
			Patches: &v1.PatchDatas{
				Patches: []*v1.PatchData{{
					Path:     FieldScheme,
					Value:    templateIns.Scheme().Raw(),
					Operator: xjson.OpReplace.String()},
				}}}}
	err = r.dispatcher.Dispatch(ctx, ev)
	return errors.Wrap(err, "On Template Changed")
}

type tsData struct {
	TS    int64       `json:"ts"`
	Value interface{} `json:"value"`
}

type tsDevice struct {
	TS     int64                  `json:"ts"`
	Values map[string]interface{} `json:"values"`
}

func adjustTSData(bytes []byte) (dataAdjust []byte) {
	// tsDevice1 no ts
	tsDevice1 := make(map[string]interface{})
	err := json.Unmarshal(bytes, &tsDevice1)
	if err == nil && len(tsDevice1) > 0 {
		tsDeviceAdjustData := make(map[string]*tsData)
		for k, v := range tsDevice1 {
			switch v.(type) {
			case map[string]interface{}:
				goto dataType2
			default:
			}
			tsDeviceAdjustData[k] = &tsData{TS: time.Now().UnixMilli(), Value: v}
		}
		dataAdjust, _ = json.Marshal(tsDeviceAdjustData)
		return
	}

	// tsDevice2 has ts
dataType2:
	tsDevice2 := tsDevice{}
	err = json.Unmarshal(bytes, &tsDevice2)
	if err == nil && tsDevice2.TS != 0 {
		tsDeviceAdjustData := make(map[string]*tsData)
		for k, v := range tsDevice2.Values {
			tsDeviceAdjustData[k] = &tsData{TS: tsDevice2.TS, Value: v}
		}
		dataAdjust, _ = json.Marshal(tsDeviceAdjustData)
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
		return
	}
	log.Error("ts data adjust error", zap.Error(err))
	return dataAdjust
}

func (r *Runtime) handleRawData(ctx context.Context, feed *Feed) *Feed {
	log.L().Debug("handle RawData", zfield.Eid(feed.EntityID))

	// match properties.rawData.
	for _, patch := range feed.Patches {
		if FieldRawData == patch.Path {
			// attempt extract rawData.
			prefix := patch.Value.Get("type").String()

			if prefix == rawDataRawType {
				return feed
			}

			values := patch.Value.Get("values").String()
			bytes, err := base64.StdEncoding.DecodeString(values)
			if nil != err {
				log.L().Warn("attempt extract RawData", zfield.Eid(feed.EntityID),
					zfield.Reason(err.Error()), zfield.Value(patch.Value.String()))
				return feed
			}

			log.L().Debug("extract RawData successful", zfield.Eid(feed.EntityID),
				zap.Any("raw", patch.Value.String()), zap.String("value", string(bytes)))

			if prefix == rawDataTelemetryType {
				bytes = adjustTSData(bytes)
			}

			path := strings.Join([]string{FieldProperties, prefix}, ".")
			feed.Patches = append(feed.Patches, Patch{
				Path:  path,
				Value: tdtl.New(bytes),
				Op:    xjson.OpMerge,
			})
			return feed
		}
	}

	return feed
}

func (r *Runtime) AppendExpression(exprInfo ExpressionInfo) {
	log.L().Debug("append expression into runtime",
		zfield.ID(exprInfo.ID), zfield.Eid(exprInfo.EntityID),
		zfield.Owner(exprInfo.Owner), zfield.Expr(exprInfo.Expression.Expression))

	// remove expression if exists.
	if exprOld, exists := r.getExpr(exprInfo.ID); exists {
		// remove sub-endpoint from sub-tree.
		for _, item := range exprOld.subEndpoints {
			r.subTree.Remove(item.WildcardPath(), &item)
		}

		// remove eval-endpoint from eval-tree.
		for _, item := range exprOld.evalEndpoints {
			r.evalTree.Remove(item.WildcardPath(), &item)
		}
	}

	// cache expression info.
	r.setExpr(exprInfo)

	// mount sub-endpoint to sub-tree.
	for _, item := range exprInfo.subEndpoints {
		r.subTree.Add(item.WildcardPath(), &item)
	}
	// mount eval-endpoint to eval-tree.
	for _, item := range exprInfo.evalEndpoints {
		r.evalTree.Add(item.WildcardPath(), &item)
	}

	r.initializeExpression(context.TODO(), exprInfo)
}

func (r *Runtime) RemoveExpression(exprID string) {
	// remove expression if exists.
	if exprInfo, exists := r.getExpr(exprID); exists {
		log.L().Debug("remove expression from runtime",
			zfield.ID(exprInfo.ID), zfield.Eid(exprInfo.EntityID),
			zfield.Owner(exprInfo.Owner), zfield.Expr(exprInfo.Expression.Expression))

		// remove sub-endpoint from sub-tree.
		for _, item := range exprInfo.subEndpoints {
			r.subTree.Remove(item.WildcardPath(), &item)
		}

		// remove eval-endpoint from eval-tree.
		for _, item := range exprInfo.evalEndpoints {
			r.evalTree.Remove(item.WildcardPath(), &item)
		}
	}
}

func (r *Runtime) initializeExpression(ctx context.Context, expr ExpressionInfo) {
	if mapper.VersionInited != expr.version {
		return
	}

	log.L().Info("initialize expression", zfield.ID(r.id),
		zfield.Eid(expr.EntityID), zfield.ID(expr.ID), zfield.Value(expr.Expression))

	if len(expr.evalEndpoints) > 0 {
		// TODO: 解决 Cache 消息 先于 mapper 初始化, 需要深入思考原因.
		patches := []*v1.PatchData{}
		result, err := r.evalExpression(ctx, expr.Expression)
		if nil != err {
			log.L().Error("eval expression",
				zfield.Eid(expr.EntityID), zfield.ID(expr.ID),
				zfield.Expr(expr.Expression.Expression))
			return
		} else if nil == result {
			log.L().Warn("eval expression, empty result.",
				zfield.Eid(expr.EntityID), zfield.ID(expr.ID),
				zfield.Expr(expr.Expression.Expression))
			return
		}

		log.L().Debug("eval expression", zfield.Expr(expr.Expression.Expression),
			zfield.Eid(expr.EntityID), zfield.ID(expr.ID), zfield.Owner(expr.Owner))

		patches = append(
			patches,
			&v1.PatchData{
				Operator: xjson.OpReplace.String(),
				Path:     expr.Expression.Path,
				Value:    result.Raw(),
			})

		// 2. dispatch.send() .
		r.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
			Id:        util.IG().EvID(),
			Timestamp: time.Now().UnixNano(),
			Metadata: map[string]string{
				v1.MetaType:     string(v1.ETEntity),
				v1.MetaBorn:     "initializeExpression",
				v1.MetaEntityID: expr.EntityID},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: patches,
				}},
		})
	} else {
		patches := map[string][]*v1.PatchData{}
		for _, subEnd := range expr.subEndpoints {
			var err error
			var val tdtl.Node
			var state Entity
			item := mapper.NewWatchKey(subEnd.path)
			if state, err = r.LoadEntity(item.EntityID); nil != err {
				log.L().Warn("load entity", zap.Error(err), zfield.Eid(item.EntityID))
				continue
			}

			operator := xjson.OpReplace.String()
			path := item.PropertyKey
			if path == "*" {
				// TODO: 现阶段 TQL 仅仅支持 eid.* .
				path = FieldProperties
				operator = xjson.OpMerge.String()
			}

			if val = state.Get(path); nil != val.Error() {
				log.L().Warn("get entity property", zap.Error(val.Error()), zfield.Eid(item.EntityID))
				continue
			}

			patches[item.EntityID] =
				append(patches[item.EntityID],
					&v1.PatchData{
						Path:     path,
						Value:    val.Raw(),
						Operator: operator})
		}

		// handle subscribe, dispatch entity state.
		for entityID, patch := range patches {
			r.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
				Id:        util.IG().EvID(),
				Timestamp: time.Now().UnixNano(),
				Metadata: map[string]string{
					v1.MetaType:     string(v1.ETCache),
					v1.MetaBorn:     "initializeExpression",
					v1.MetaEntityID: expr.EntityID,
					v1.MetaSender:   entityID},
				Data: &v1.ProtoEvent_Patches{
					Patches: &v1.PatchDatas{
						Patches: patch,
					}},
			})
		}
	}
}

func (r *Runtime) LoadEntity(id string) (Entity, error) {
	r.lock.Lock()
	if state, ok := r.entities[id]; ok {
		r.lock.Unlock()
		return state, nil
	}
	r.lock.Unlock()

	// load from state storage.
	jsonData, err := r.repository.GetEntity(context.TODO(), id)
	if nil != err {
		log.L().Warn("load entity from state storage",
			zfield.Eid(id), zfield.Reason(err.Error()))
		return nil, errors.Wrap(err, "load entity")
	}

	// create entity instance.
	en, err := NewEntity(id, jsonData)
	if nil != err {
		log.L().Warn("create entity instance",
			zfield.Eid(id), zfield.Reason(err.Error()))
		return nil, errors.Wrap(err, "create entity instance")
	}

	r.lock.Lock()
	r.entities[id] = en
	r.lock.Unlock()

	return en, nil
}

func conv(patches []*v1.PatchData) []Patch {
	res := make([]Patch, 0)
	for _, patch := range patches {
		res = append(res, Patch{
			Op:    xjson.NewPatchOp(patch.Operator),
			Path:  patch.Path,
			Value: tdtl.New(patch.Value),
		})
	}
	return res
}

func (r *Runtime) getExpr(id string) (ExpressionInfo, bool) {
	r.mlock.RLock()
	defer r.mlock.RUnlock()
	exprInfo, has := r.expressions[id]
	return exprInfo, has
}

func (r *Runtime) setExpr(exprInfo ExpressionInfo) {
	r.mlock.Lock()
	defer r.mlock.Unlock()
	r.expressions[exprInfo.ID] = exprInfo
}
