package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper/expression"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
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
