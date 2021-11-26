package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatch(t *testing.T) {
	var err error
	dest := NewNode(`{"temp":20}`)
	dest, err = Patch(dest, NewNode(22), "temp", PatchOpReplace)
	assert.Nil(t, err)
	assert.Equal(t, `{"temp":22}`, dest.String())

	dest, err = Patch(dest, NewNode("555"), "temp", PatchOpReplace)
	assert.Nil(t, err)
	assert.Equal(t, `{"temp":"555"}`, dest.String())

	dest, err = Patch(dest, NewNode("555"), "append", PatchOpAdd)
	assert.Nil(t, err)
	assert.Equal(t, `{"temp":"555","append":["555"]}`, dest.String())

	dest, err = Patch(dest, NewNode("555"), "append[0]", PatchOpRemove)
	assert.Nil(t, err)
	assert.Equal(t, `{"temp":"555","append":[]}`, dest.String())

	dest, err = Patch(dest, NewNode(map[string]interface{}{"property1": 12345}), "append", PatchOpAdd)
	assert.Nil(t, err)
	assert.Equal(t, "{\"temp\":\"555\",\"append\":[{\"property1\":12345}]}", dest.String())

	dest, err = Patch(dest, NewNode("test"), "append", PatchOpAdd)
	assert.Nil(t, err)
	assert.Equal(t, `{"temp":"555","append":[{"property1":12345},"test"]}`, dest.String())
}

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
