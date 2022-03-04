package runtime2

import (
	"context"
	"fmt"
	"strings"
)

type SourceConf struct {
	Topic      string
	Brokers    []string
	Partitions []int32
}

type RuntimeConfig struct {
	Source SourceConf
}

func NewContainer(ctx context.Context) *Container {
	return &Container{}
}

type Runtime struct {
	containers map[string]*Container
	dispatch   Dispatch
	dao        Dao

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRuntime(ctx context.Context, d Dao, dispatcher Dispatch) *Runtime {
	ctx, cacel := context.WithCancel(ctx)
	return &Runtime{
		dao:        d,
		ctx:        ctx,
		cancel:     cacel,
		dispatch:   dispatcher,
		containers: make(map[string]*Container),
	}
}

func (r *Runtime) Start(cfg RuntimeConfig) error {

	var source = cfg.Source
	var sourceUrls []string
	// reconstruct configs.
	for _, partition := range source.Partitions {
		sourceUrls = append(sourceUrls, fmt.Sprintf("partition://%s/%s/%d",
			strings.Join(source.Brokers, ";"), source.Topic, partition))
	}

	for _, sourceUrl := range sourceUrls {
		sourceIns := NewPartitionSource(r.ctx, sourceUrl)

		// new container.
		container := NewContainer(r.ctx)

		// consume source.
		err := sourceIns.StartReceiver(r.ctx, container.DeliveredEvent)
		if nil != err {
			panic(err)
		}
	}

	return nil
}
