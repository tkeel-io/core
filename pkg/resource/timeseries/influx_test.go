package timeseries

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Write(t *testing.T) {
	// outer := newInflux()
	// outer.Init(resource.Metadata{
	// 	Name: "influxdb",
	// 	Properties: map[string]string{
	// 		"org":    "yunify",
	// 		"bucket": "entity",
	// 		"url":    "http://localhost:8086",
	// 		"token":  "9bUWcVwUpxbNSuhJMLbRaJxCVl8LzFV33znGx-pAXg4HUxFgWRTkRArF5Z9lMDcOn1pzzfD4dovLkkTnxuVMtg==",
	// 	},
	// })

	// num := 10000
	// t.Log("write som data, N=", num)
	// for i := 0; i < num; i++ {
	// 	_, err := outer.Write(context.Background(), &tseries.TSeriesRequest{
	// 		Raw: []string{fmt.Sprintf("mem,host=host1 used_percent=%f %d", 40.0, time.Now().Unix())},
	// 	})
	// 	if nil != err {
	// 		t.Log("write influx failed", err)
	// 	}
	// }
}

func TestFluxQuery(t *testing.T) {
	tests := []struct {
		name string
		q    FluxQuery
		want string
	}{
		{"bucket", FluxQuery{FromBucket: "test"}, `from(bucket: "test")`},
		{"bucket and range", FluxQuery{FromBucket: "test", RangeOpt: FluxRangeOption("now()", "now()")},
			`from(bucket: "test")
|> range(start: now(), stop: now())`},
		{"bucket and range and filter", FluxQuery{FromBucket: "test", RangeOpt: FluxRangeOption("now()", "now()"), FilterOpts: []FluxOption{FluxMeasurementFilterOption("and", "_test")}},
			`from(bucket: "test")
|> range(start: now(), stop: now())
|> filter(fn: (r) =>
r._measurement == "_test"
)`},
		{"bucket and range and multiple filter",
			FluxQuery{FromBucket: "test", RangeOpt: FluxRangeOption("now()", "now()"),
				FilterOpts: []FluxOption{FluxMeasurementFilterOption("and", "_test"), FluxFiledFilterOption("and", "_field"), FluxTagFilterOption("or", "myTag")}},
			`from(bucket: "test")
|> range(start: now(), stop: now())
|> filter(fn: (r) =>
r._measurement == "_test" and
r._field == "_field" or
r.tag == "myTag"
)`},
		{"bucket and range and multiple filter and sort",
			FluxQuery{FromBucket: "test", RangeOpt: FluxRangeOption("now()", "now()"),
				FilterOpts: []FluxOption{FluxMeasurementFilterOption("and", "_test"), FluxFiledFilterOption("and", "_field")},
				SortOpt:    FluxSortOption([]string{"myColumn1", "myColumn2"})},
			`from(bucket: "test")
|> range(start: now(), stop: now())
|> filter(fn: (r) =>
r._measurement == "_test" and
r._field == "_field"
)
|> sort(columns: ["myColumn1","myColumn2"])`},
		{"bucket and group with mode by", FluxQuery{FromBucket: "test", GroupOpt: FluxGroupQueryOption([]string{"column1", "column2"}, "by")},
			`from(bucket: "test")
|> group(columns:["column1","column2"], mode:"by")`},
		{"bucket and group without mode", FluxQuery{FromBucket: "test", GroupOpt: FluxGroupQueryOption([]string{"column1", "column2"})},
			`from(bucket: "test")
|> group(columns:["column1","column2"])`},
		{"bucket and group without mode and limit", FluxQuery{FromBucket: "test", GroupOpt: FluxGroupQueryOption([]string{"column1", "column2"}), LimitOpt: FluxLimitOption(5)},
			`from(bucket: "test")
|> group(columns:["column1","column2"])
|> limit(n: 5)`},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, test.q.ToRawQuery())
	}
}
