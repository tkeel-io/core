package runtime

import (
	"context"
	"time"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

func (m *Manager) repo() repository.IRepository {
	return m.resourceManager.Repository()
}

// initialize runtime environments.
func (m *Manager) initialize() {
	elapsedTime := util.NewElapsed()
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	revision := m.repo().GetLastRevision(context.Background())
	log.Info("initialize actor manager, mapper loadding...")
	m.repo().RangeMapper(ctx, revision, func(mappers []dao.Mapper) {
		for _, info := range m.actorEnv.StoreMappers(mappers) {
			log.Debug("load actor", zfield.ID(info.ID), zfield.Name(info.Name), zfield.TQL(info.TQL),
				zfield.Eid(info.EntityID), zfield.Type(info.EntityType), zfield.Desc(info.Description))
			// load actor when environment initialized.
			if err := m.loadActor(context.Background(), info.EntityType, info.EntityID); nil != err {
				log.Error("load actor", zfield.Eid(info.EntityID), zfield.Type(info.EntityType))
			}
		}
	})

	log.Debug("runtime.Environment initialized", zfield.Elapsedms(elapsedTime.Elapsed()))
}

// watchResource watch resources.
func (m *Manager) watchResource() {
	m.repo().WatchMapper(context.Background(),
		m.repo().GetLastRevision(context.Background()),
		func(et dao.EnventType, mp dao.Mapper) {
			effects, _ := m.actorEnv.OnMapperChanged(et, mp)
			m.reloadActor(effects)
		})
}
