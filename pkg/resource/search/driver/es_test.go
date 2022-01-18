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

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/tkeel-io/core/api/core/v1"
)

func Test_condition2boolQuery(t *testing.T) {

}

func Test_defaultPage(t *testing.T) {
	page := &pb.Pager{Offset: 20}

	defaultPage(page)
	assert.Equal(t, int64(10), page.Limit)
}
