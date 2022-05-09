package runtime

import (
	"context"
	"fmt"
	go_restful "github.com/emicklei/go-restful"
	"github.com/tkeel-io/core/pkg/util/path"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/mapper/expression"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	xkafka "github.com/tkeel-io/core/pkg/util/kafka"
	"github.com/tkeel-io/kit/log"
)

type NodeConf struct {
	Sources []string
}

type Node struct {
	runtimes        map[string]*Runtime
	queues          map[string]*xkafka.Pubsub
	dispatch        dispatch.Dispatcher
	resourceManager types.ResourceManager
	revision        int64
	lock            sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
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
		queues:          make(map[string]*xkafka.Pubsub),
	}
}

// Start Node
//1. 创建 KafkaSource & runtime
//2. list resource
//3. watch resource
//4. start KafkaReceived.
func (n *Node) Start(cfg NodeConf) error {
	log.L().Info("start node...")

	//1. 创建 KafkaSource & runtime
	var err error
	var sourceIns *xkafka.Pubsub
	for index := range cfg.Sources {
		if sourceIns, err = xkafka.NewKafkaPubsub(cfg.Sources[index]); nil != err {
			return errors.Wrap(err, "create source instance")
		}
		runtimeID := sourceIns.ID()
		n.queues[runtimeID] = sourceIns
		// create runtime instance.
		log.L().Info("create runtime instance",
			logf.ID(runtimeID), logf.Source(cfg.Sources[index]))
		entityResouce := EntityResource{FlushHandler: n.FlushEntity, RemoveHandler: n.RemoveEntity}
		runtime := NewRuntime(n.ctx, entityResouce, runtimeID, n.dispatch, n.resourceManager.Repo())
		n.runtimes[runtimeID] = runtime
		placement.Global().Append(placement.Info{ID: sourceIns.ID(), Flag: true})
	}

	//2. list resource
	var elapsed util.ElapsedTime
	n.listMetadata()

	//3. watch resource
	n.watchMetadata()

	//4. start KafkaReceived
	for _, queue := range n.queues {
		if err = queue.Received(n.ctx, n); nil != err {
			return errors.Wrap(err, "consume source")
		}
	}
	// watch metadata.
	log.L().Debug("start node completed", logf.Elapsedms(elapsed.ElapsedMilli()))
	//
	//for index := range cfg.Sources {
	//	var err error
	//	var sourceIns *xkafka.Pubsub
	//	if sourceIns, err = xkafka.NewKafkaPubsub(cfg.Sources[index]); nil != err {
	//		return errors.Wrap(err, "create source instance")
	//	} else if err = sourceIns.Received(n.ctx, n); nil != err {
	//		return errors.Wrap(err, "consume source")
	//	}
	//
	//	rid := sourceIns.ID()
	//	// create runtime instance.
	//	log.L().Info("create runtime instance",
	//		logf.ID(rid), logf.Source(cfg.Sources[index]))
	//
	//	entityResouce := EntityResource{FlushHandler: n.FlushEntity, RemoveHandler: n.RemoveEntity}
	//	rt := NewRuntime(n.ctx, entityResouce, rid, n.dispatch, n.resourceManager.Repo())
	//	for _, expr := range n.expressions {
	//		exprInfos, err := parseExpression(expr.Expression, 1)
	//		if nil != err {
	//			log.L().Error("parse expression", logf.Eid(expr.EntityID),
	//				logf.Expr(expr.Expression.Expression), logf.Desc(expr.Description),
	//				logf.Mid(expr.Path), logf.Owner(expr.Owner), logf.Name(expr.Name), logf.Error(err))
	//			continue
	//		}
	//
	//		if exprIns, has := exprInfos[rt.ID()]; has {
	//			rt.AppendExpression(*exprIns)
	//		}
	//	}
	//
	//	n.runtimes[rid] = rt
	//	placement.Global().Append(placement.Info{ID: sourceIns.ID(), Flag: true})
	//}
	//
	//

	return nil
}

func (n *Node) HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	rid := msg.Topic
	if _, has := n.runtimes[rid]; !has {
		log.L().Error("runtime instance not exists.", logf.ID(rid),
			logf.Any("header", msg.Headers), logf.Message(string(msg.Value)))
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
			log.L().Debug("sync expression", logf.Eid(expr.EntityID),
				logf.Expr(expr.Expression), logf.Desc(expr.Description),
				logf.Mid(expr.Path), logf.Owner(expr.Owner), logf.Name(expr.Name))

			// cache for node.
			exprInfo := newExprInfo(expr)
			exprInfos, err := parseExpression(exprInfo.Expression, 1)
			if nil != err {
				log.L().Error("parse expression", logf.Eid(expr.EntityID),
					logf.Expr(expr.Expression), logf.Desc(expr.Description),
					logf.Mid(expr.Path), logf.Owner(expr.Owner), logf.Name(expr.Name), logf.Error(err))
				continue
			}
			for runtimeID, exprIns := range exprInfos {
				runtime, ok := n.runtimes[runtimeID]
				if ok {
					runtime.AppendExpression(*exprIns)
				}
			}
		}
	})

	repo.RangeSubscription(ctx, n.revision, func(subscriptions []*repository.Subscription) {
		// 将mapper加入每一个 runtime.
		for _, sub := range subscriptions {
			log.L().Debug("sync subscription", logf.String("subID", sub.ID), logf.Owner(sub.Owner))
			entityID := sub.SourceEntityID
			runtimeInfo := placement.Global().Select(entityID)
			runtime, ok := n.runtimes[runtimeInfo.ID]
			if ok {
				_, ok := runtime.entitySubscriptions[entityID]
				if !ok {
					runtime.entitySubscriptions[entityID] = make(map[string]*repository.Subscription)
				}
				runtime.entitySubscriptions[entityID][sub.ID] = sub
			}
		}
	})
	log.L().Debug("runtime.Environment initialized", logf.Elapsedms(elapsedTime.ElapsedMilli()))
}

// watchResource watch resources.
func (n *Node) watchMetadata() {
	repo := n.resourceManager.Repo()
	go repo.WatchExpression(context.Background(), n.revision,
		func(et dao.EnventType, expr repository.Expression) {
			switch et {
			case dao.DELETE:
				exprInfo := newExprInfo(&expr)
				log.L().Debug("sync DELETE expression", logf.Eid(expr.EntityID),
					logf.Expr(expr.Expression), logf.Desc(expr.Description),
					logf.Mid(expr.Path), logf.Owner(expr.Owner), logf.Name(expr.Name))

				// remove mapper from all runtime.
				for _, rt := range n.runtimes {
					rt.RemoveExpression(exprInfo.ID)
				}
			case dao.PUT:
				exprInfo := newExprInfo(&expr)
				log.L().Debug("sync expression", logf.Eid(expr.EntityID),
					logf.Expr(expr.Expression), logf.Desc(expr.Description),
					logf.Mid(expr.Path), logf.Owner(expr.Owner), logf.Name(expr.Name))

				exprInfos, err := parseExpression(exprInfo.Expression, 0)
				if nil != err {
					log.L().Error("parse expression", logf.Eid(expr.EntityID),
						logf.Expr(expr.Expression), logf.Desc(expr.Description),
						logf.Mid(expr.Path), logf.Owner(expr.Owner), logf.Name(expr.Name), logf.Error(err))
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

	go repo.WatchSubscription(context.Background(), n.revision,
		func(et dao.EnventType, sub *repository.Subscription) {
			switch et {
			case dao.DELETE:
				log.L().Debug("sync DELETE Subscription", logf.String("subID", sub.ID), logf.Owner(sub.Owner))
				entityID := sub.SourceEntityID
				runtimeInfo := placement.Global().Select(entityID)
				runtime, ok := n.runtimes[runtimeInfo.ID]
				if ok {
					if subscription, ok := runtime.entitySubscriptions[sub.SourceEntityID]; ok {
						delete(subscription, sub.ID)
					}
				}
			case dao.PUT:
				log.L().Debug("sync PUT Subscription", logf.String("subID", sub.ID), logf.Owner(sub.Owner))
				entityID := sub.SourceEntityID
				runtimeInfo := placement.Global().Select(entityID)
				runtime, ok := n.runtimes[runtimeInfo.ID]
				if ok {
					_, ok := runtime.entitySubscriptions[entityID]
					if !ok {
						runtime.entitySubscriptions[entityID] = make(map[string]*repository.Subscription)
					}
					runtime.entitySubscriptions[entityID][sub.ID] = sub
				}
			default:
				log.L().Error("watch metadata changed, invalid event type")
			}
		})
}

func parseExpression(expr repository.Expression, version int) (map[string]*ExpressionInfo, error) {
	exprIns, err := expression.NewExpr(expr.Expression, nil)
	if nil != err {
		return nil, errors.Wrap(err, "parse expression")
	}

	targetRuntimeInfo := placement.Global().Select(expr.EntityID)
	exprInfos := map[string]*ExpressionInfo{
		targetRuntimeInfo.ID: {
			version:    version,
			Expression: expr,
		}}

	for sourceEntityID, paths := range exprIns.Sources() {
		sourceRuntimeInfo := placement.Global().Select(sourceEntityID)
		if _, has := exprInfos[sourceRuntimeInfo.ID]; !has {
			exprInfos[sourceRuntimeInfo.ID] = &ExpressionInfo{
				version:    version,
				Expression: expr,
			}
		}

		for _, path := range paths {
			// construct sub endpoint.
			if sourceEntityID != expr.EntityID {
				exprInfos[sourceRuntimeInfo.ID].subEndpoints =
					append(exprInfos[sourceRuntimeInfo.ID].subEndpoints,
						newSubEnd(path, expr.EntityID, expr.ID, targetRuntimeInfo.ID))
			}

			// construct eval endpoint.
			if repository.ExprTypeEval == expr.Type {
				exprInfos[targetRuntimeInfo.ID].evalEndpoints =
					append(exprInfos[targetRuntimeInfo.ID].evalEndpoints,
						newEvalEnd(path, expr.EntityID, expr.ID))
			} else if repository.ExprTypeSub == expr.Type {
				exprInfos[targetRuntimeInfo.ID].subEndpoints =
					append(exprInfos[targetRuntimeInfo.ID].subEndpoints,
						newSubEnd(path, expr.EntityID, expr.ID, targetRuntimeInfo.ID))
			}
		}
	}

	return exprInfos, nil
}

// exprKey return unique expression identifier.
func exprKey(expr *repository.Expression) string { //nolint
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

func (n *Node) Debug(req *go_restful.Request, resp *go_restful.Response) {
	fmt.Println("Debug", req.Request.URL)
	action := req.Request.URL.Query().Get("action")
	switch action {
	case "nodelist":
		ret := []string{}
		for k, _ := range n.runtimes {
			ret = append(ret, k)
		}
		resp.WriteAsJson(strings.Join(ret, "|"))
	case "subtree":
		rid := req.Request.URL.Query().Get("rid")
		ret := n.runtimes[rid]
		resp.Write([]byte(ret.subTree.String()))
	case "eveltree":
		rid := req.Request.URL.Query().Get("rid")
		ret := n.runtimes[rid]
		resp.Write([]byte(ret.evalTree.String()))
	case "sub":
		rid := req.Request.URL.Query().Get("rid")
		entityID := req.Request.URL.Query().Get("entityID")
		changePath := req.Request.URL.Query().Get("changePath")
		rt, ok := n.runtimes[rid]
		if ok {
			ret := rt.subTree.MatchPrefix(path.FmtWatchKey(entityID, changePath))
			resp.WriteAsJson(ret)
		} else {
			resp.WriteErrorString(501, "runtime <"+rid+"> not found")
		}
	}
}
