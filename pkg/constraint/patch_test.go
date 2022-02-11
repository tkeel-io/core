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

package constraint

import (
	"testing"

	"github.com/tkeel-io/collectjs"
)

// func TestPatch(t *testing.T) {
// 	var err error
// 	var dest Node = JSONNode(`{"temp":20}`)
// 	dest, err = Patch(dest, NewNode(22), "temp", PatchOpReplace)
// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"temp":22}`, dest.String())

// 	dest, err = Patch(dest, NewNode("555"), "temp", PatchOpReplace)
// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"temp":"555"}`, dest.String())

// 	dest, err = Patch(dest, NewNode("555"), "append", PatchOpAdd)
// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"temp":"555","append":["555"]}`, dest.String())

// 	dest, err = Patch(dest, NewNode("555"), "append[0]", PatchOpRemove)
// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"temp":"555","append":[]}`, dest.String())

// 	dest, err = Patch(dest, NewNode(map[string]interface{}{"property1": 12345}), "append", PatchOpAdd)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "{\"temp\":\"555\",\"append\":[{\"property1\":12345}]}", dest.String())

// 	dest, err = Patch(dest, NewNode("test"), "append", PatchOpAdd)
// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"temp":"555","append":[{"property1":12345},"test"]}`, dest.String())
// }

func BenchmarkPatch1(b *testing.B) {
	raw := JSONNode(`{"temp":"555"}`)
	//	expect := `{"temp":"555","append":[{"property1":9999},"test"]}`
	for n := 0; n < b.N; n++ {
		Patch(raw, NewNode(9999), "temp", PatchOpReplace)
	}
}

func BenchmarkPatch2(b *testing.B) {
	raw := JSONNode(`{"temp":"555","append":[{"property1":12345},"test"]}`)
	//	expect := `{"temp":"555","append":[{"property1":9999},"test"]}`
	for n := 0; n < b.N; n++ {
		Patch(raw, NewNode(9999), "append[0].property1", PatchOpRemove)
	}
}

func TestCollectEmptyPath(t *testing.T) {
	result, _ := collectjs.Append([]byte("[]"), "", []byte(`20`))
	_, _ = collectjs.Set(result, "[0]", []byte(`2220`))
	result, _ = collectjs.Set([]byte(`{}`), "age", []byte(`2220`))
	t.Log("result: ", string(result))
}
