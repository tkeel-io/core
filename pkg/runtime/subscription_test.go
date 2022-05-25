package runtime

import (
	"strings"
	"testing"

	"github.com/tkeel-io/core/pkg/repository"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/tdtl"
)

func Test_makeSubData(t *testing.T) {
	cc := tdtl.New(`{}`)
	cc.Set("properties.temps.temp", tdtl.IntNode(20))
	cc.Set("properties.temps.temp2", tdtl.IntNode(20))
	cc.Set("properties.metrics.cpu.value", tdtl.FloatNode(0.78))
	cc.Set("properties.metrics.cpu1.value", tdtl.FloatNode(0.78))
	sub := repository.Subscription{}
	sub.ID = "subID"
	sub.Owner = "owner"
	t.Log(strings.TrimSuffix("aaa", ".*"))
	t.Log(strings.TrimSuffix("aaa.*", ".*"))
	sub.SourceEntityPaths = []string{"properties1.*"}
	state := cc.Raw()
	bytes := makeSubData(&Feed{
		State: state,
		Changes: []Patch{
			{
				Op:    xjson.OpMerge,
				Path:  "properties.temps",
				Value: tdtl.New(`{"temp":20}`),
			},
			{
				Op:    xjson.OpReplace,
				Path:  "properties.metrics.cpu.value",
				Value: tdtl.New(``),
			},
		},
	}, &sub)
	t.Log("payload: ", string(bytes))
}
