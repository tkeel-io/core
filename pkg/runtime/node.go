package runtime

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper/expression"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	xkafka "github.com/tkeel-io/core/pkg/util/kafka"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

type NodeConf struct {
	Sources []string
}

type Node struct {
	runtimes        map[string]*Runtime
	dispatch        dispatch.Dispatcher
	resourceManager types.ResourceManager
	expressions     map[string]ExpressionInfo
	revision        int64

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewNode(ctx context.Context, resourceManager types.ResourceManager, dispatcher dispatch.Dispatcher) *Node {
	ctx, cacel := context.WithCancel(ctx)
	return &Node{
		ctx:             ctx,
		cancel:          cacel,
		lock:            sync.RWMutex{},
		dispatch:        dispatcher,
		resourceManager: resourceManager,
		runtimes:        make(map[string]*Runtime),
		expressions:     make(map[string]ExpressionInfo),
	}
}

func (n *Node) Start(cfg NodeConf) error {
	log.L().Info("start node...")

	var elapsed util.ElapsedTime
	n.listMetadata()
	for index := range cfg.Sources {
		var err error
		var sourceIns *xkafka.Pubsub
		if sourceIns, err = xkafka.NewKafkaPubsub(cfg.Sources[index]); nil != err {
			return errors.Wrap(err, "create source instance")
		} else if err = sourceIns.Received(n.ctx, n); nil != err {
			return errors.Wrap(err, "consume source")
		}

		rid := sourceIns.ID()
		// create runtime instance.
		log.L().Info("create runtime instance",
			zfield.ID(rid), zfield.Source(cfg.Sources[index]))

		entityResouce := EntityResource{FlushHandler: n.FlushEntity, RemoveHandler: n.RemoveEntity}
		rt := NewRuntime(n.ctx, entityResouce, rid, n.dispatch, n.resourceManager.Repo())
		for _, expr := range n.expressions {
			exprInfos, err := parseExpression(expr.Expression, 1)
			if nil != err {
				log.L().Error("parse expression", zfield.Eid(expr.EntityID),
					zfield.Expr(expr.Expression.Expression), zfield.Desc(expr.Description),
					zfield.Mid(expr.Path), zfield.Owner(expr.Owner), zfield.Name(expr.Name), zap.Error(err))
				continue
			}

			if exprIns, has := exprInfos[rt.ID()]; has {
				rt.AppendExpression(*exprIns)
			}
		}

		n.runtimes[rid] = rt
		placement.Global().Append(placement.Info{ID: sourceIns.ID(), Flag: true})
	}

	// release expressions.
	n.expressions = nil

	// watch metadata.
	go n.watchMetadata()
	log.L().Debug("start node completed", zfield.Elapsedms(elapsed.ElapsedMilli()))

	return nil
}

func (n *Node) HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	rid := msg.Topic
	if _, has := n.runtimes[rid]; !has {
		log.L().Error("runtime instance not exists.", zfield.ID(rid),
			zap.Any("header", msg.Headers), zfield.Message(string(msg.Value)))
		return xerrors.ErrRuntimeNotExists
	}

	// load runtime spec.
	rt := n.runtimes[rid]
	rt.DeliveredEvent(context.Background(), msg)
	return nil
}

// initialize runtime environments.
func (n *Node) listMetadata() {
	elapsedTime := util.NewElapsed()
	ctx, cancel := context.WithTimeout(n.ctx, 30*time.Second)
	defer cancel()

	repo := n.resourceManager.Repo()
	n.revision = repo.GetLastRevision(context.Background())
	log.L().Info("initialize actor manager, mapper loadding...")
	repo.RangeExpression(ctx, n.revision, func(expressions []*repository.Expression) {
		// 将mapper加入每一个 runtime.
		for _, expr := range expressions {
			log.L().Debug("sync expression", zfield.Eid(expr.EntityID),
				zfield.Expr(expr.Expression), zfield.Desc(expr.Description),
				zfield.Mid(expr.Path), zfield.Owner(expr.Owner), zfield.Name(expr.Name))

			// cache for node.
			n.expressions[exprKey(expr)] = newExprInfo(expr)
		}
	})

	log.L().Debug("runtime.Environment initialized", zfield.Elapsedms(elapsedTime.ElapsedMilli()))
}

// watchResource watch resources.
func (n *Node) watchMetadata() {
	repo := n.resourceManager.Repo()
	repo.WatchExpression(context.Background(), n.revision,
		func(et dao.EnventType, expr repository.Expression) {
			switch et {
			case dao.DELETE:
				exprInfo := newExprInfo(&expr)
				log.L().Debug("sync DELETE expression", zfield.Eid(expr.EntityID),
					zfield.Expr(expr.Expression), zfield.Desc(expr.Description),
					zfield.Mid(expr.Path), zfield.Owner(expr.Owner), zfield.Name(expr.Name))

				// remove mapper from all runtime.
				for _, rt := range n.runtimes {
					rt.RemoveExpression(exprInfo.ID)
				}
			case dao.PUT:
				exprInfo := newExprInfo(&expr)
				log.L().Debug("sync expression", zfield.Eid(expr.EntityID),
					zfield.Expr(expr.Expression), zfield.Desc(expr.Description),
					zfield.Mid(expr.Path), zfield.Owner(expr.Owner), zfield.Name(expr.Name))

				exprInfos, err := parseExpression(exprInfo.Expression, 0)
				if nil != err {
					log.L().Error("parse expression", zfield.Eid(expr.EntityID),
						zfield.Expr(expr.Expression), zfield.Desc(expr.Description),
						zfield.Mid(expr.Path), zfield.Owner(expr.Owner), zfield.Name(expr.Name), zap.Error(err))
					return
				}

				// delivery expression.
				for rtID, exprItem := range exprInfos {
					if rt, has := n.runtimes[rtID]; has {
						rt.AppendExpression(*exprItem)
					}
				}
			default:
				log.L().Error("watch metadata changed, invalid event type")
			}
		})
}

func (n *Node) getGlobalData(en Entity) (res []byte) {
	globalData := collectjs.ByteNew([]byte(`{}`))
	globalData.Set(FieldID, en.Get(FieldID).Raw())
	globalData.Set(FieldType, en.Get(FieldType).Raw())
	globalData.Set(FieldOwner, en.Get(FieldOwner).Raw())
	globalData.Set(FieldSource, en.Get(FieldSource).Raw())
	globalData.Set(FieldTemplate, en.Get(FieldTemplate).Raw())

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
	return globalData.GetRaw()
}

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

	en.Properties()
	if err := n.flushTimeSeries(ctx, en); nil != err {
		log.L().Error("flush entity timeseries database", zap.Error(err), zfield.Eid(en.ID()))
	}
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

func parseExpression(expr repository.Expression, version int) (map[string]*ExpressionInfo, error) {
	exprIns, err := expression.NewExpr(expr.Expression, nil)
	if nil != err {
		return nil, errors.Wrap(err, "parse expression")
	}

	ownerInfo := placement.Global().Select(expr.EntityID)
	exprInfos := map[string]*ExpressionInfo{
		ownerInfo.ID: {
			version:    version,
			Expression: expr,
			isHere:     true,
		}}

	for eid, paths := range exprIns.Entities() {
		info := placement.Global().Select(eid)
		if _, has := exprInfos[info.ID]; !has {
			exprInfos[info.ID] = &ExpressionInfo{
				version:    version,
				Expression: expr,
			}
		}

		for _, path := range paths {
			// construct sub endpoint.
			if eid != expr.EntityID {
				exprInfos[info.ID].subEndpoints =
					append(exprInfos[info.ID].subEndpoints,
						newSubEnd(path, expr.EntityID, expr.ID, ownerInfo.ID))
			}

			// construct eval endpoint.
			if repository.ExprTypeEval == expr.Type {
				exprInfos[ownerInfo.ID].evalEndpoints =
					append(exprInfos[ownerInfo.ID].evalEndpoints,
						newEvalEnd(path, expr.EntityID, expr.ID))
			} else if repository.ExprTypeSub == expr.Type {
				exprInfos[ownerInfo.ID].subEndpoints =
					append(exprInfos[ownerInfo.ID].subEndpoints,
						newSubEnd(path, expr.EntityID, expr.ID, ownerInfo.ID))
			}
		}
	}

	return exprInfos, nil
}

// exprKey return unique expression identifier.
func exprKey(expr *repository.Expression) string {
	return expr.EntityID + expr.Path
}

func newExprInfo(expr *repository.Expression) ExpressionInfo {
	return ExpressionInfo{
		Expression: repository.Expression{
			ID:          expr.ID,
			Path:        expr.Path,
			Name:        expr.Name,
			Type:        expr.Type,
			Owner:       expr.Owner,
			EntityID:    expr.EntityID,
			Expression:  expr.Expression,
			Description: expr.Description,
		}}
}
