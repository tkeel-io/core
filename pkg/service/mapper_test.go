package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func Test_checkTQL(t *testing.T) {
	mIns := &dao.Mapper{
		ID:  "test",
		TQL: `insert into device123 select device234.metrics as metrics, device234.metrics.cpu as cpu`,
	}

	checkMapper(mIns)
	t.Log(mIns)

	assert.Equal(t, `insert into device123 select device234.properties.metrics as properties.metrics, device234.properties.metrics.cpu as properties.cpu`, mIns.TQL)
}
