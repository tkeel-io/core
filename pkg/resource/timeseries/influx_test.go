package timeseries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	/*

		outer := newInflux()
		err := outer.Init(resource.TimeSeriesMetadata{
			Name: "influxdb",
			Properties: map[string]string{
				"org":    "tkeel",
				"bucket": "tkeel",
				"url":    "http://localhost:8086",
				"token":  "mYpGorSK156qPQ3m_i9NdxeD3bFQhF-oRb3XbEXIWTMe0kH2HoDh_6sYa8CtwNcyKtNXYMuPoN8rjfqusQzmkQ==",
			},
		})

		assert.Nil(t, err)

		num := 1000
		t.Log("write som data, N=", num)
		for i := 0; i < num; i++ {
			t.Log("writing:", i)
			resp := outer.Write(context.Background(), &WriteRequest{
				Data: []string{fmt.Sprintf("test,host=host used_percent=%d", num)},
			})
			if resp.Error != nil {
				t.Log("write influx failed", err)
			}
		}
		t.Log("write Success!")

	*/
}

func TestQuery(t *testing.T) {
	/*

		outer := newInflux()
		err := outer.Init(resource.TimeSeriesMetadata{
			Name: "influxdb",
			Properties: map[string]string{
				"org":    "tkeel",
				"bucket": "tkeel",
				"url":    "http://localhost:8086",
				"token":  "Kg7mQM9pzPw_23XM6P8HJxiHl-D1eYbpnemTNxFkcHLp97ulUbIoqHIwGY9cWZkiGJjsbows-mWZO2ZZQOTO0Q==",
			},
		})

		assert.Nil(t, err)
		q := FluxQuery{FromBucket: "tkeel", RangeOpt: FluxRangeOption("-2h", ""), FilterOpts: []FluxOption{FluxMeasurementFilterOption("and", "test")}}
		//q := RawQueryRequest("from(bucket: \"tkeel\")\n  |> range(start: -2h)\n  |> filter(fn: (r) => r._measurement == \"mem\")")
		resp := outer.Query(context.Background(), &q)
		assert.Nil(t, resp.Error)
		fmt.Println(string(resp.Raw))

	*/
}

func TestFluxQuery(t *testing.T) {
	tests := []struct {
		name string
		q    FluxQuery
		want string
	}{
		{"bucket", FluxQuery{FromBucket: "test"}, `from(bucket: "test")`},
		{"valid test", FluxQuery{FromBucket: "tkeel", RangeOpt: FluxRangeOption("-2h", ""), FilterOpts: []FluxOption{FluxMeasurementFilterOption("and", "mem")}},
			"from(bucket: \"tkeel\")\n|> range(start: -2h)\n|> filter(fn: (r) =>\nr._measurement == \"mem\"\n)"},
		{"bucket and range", FluxQuery{FromBucket: "test", RangeOpt: FluxRangeOption("now()", "now()")},
			`from(bucket: "test")
|> range(start: now(), stop: now())`},
		{"bucket and range", FluxQuery{FromBucket: "test", RangeOpt: FluxRangeOption("now()", "")},
			`from(bucket: "test")
|> range(start: now())`},
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
