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

package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func TestEntity_GetEntity(t *testing.T) {
	// NewAPIManager(context.Background(), nil, nil).
}

func Test_checkTQL(t *testing.T) {
	mIns := &dao.Mapper{
		ID:  "test",
		TQL: `insert into iotd-098cafe6-821f-411d-8f84-3b4de355b5b7_core-broker-0 select iotd-098cafe6-821f-411d-8f84-3b4de355b5b7.*`,
	}

	checkMapper(mIns)
	t.Log(mIns)

	assert.Equal(t, `insert into device123 select device234.properties.metrics as properties.metrics, device234.properties.metrics.cpu as properties.cpu`, mIns.TQL)
}
