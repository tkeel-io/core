package timeseries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tkeel-io/core/pkg/resource"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

var EngineInflux Engine = "influxdb"

// Influx allows writing and reading InfluxDB.
type Influx struct {
	cfg      *InfluxConfig
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
}

type InfluxConfig struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Org    string `json:"org"`
	Bucket string `json:"bucket"`
}

// NewInflux returns a new kafka binding instance.
func newInflux() Actuator {
	return &Influx{}
}

// Init does metadata parsing and connection establishment.
func (i *Influx) Init(metadata resource.TimeSeriesMetadata) error {
	influxMeta, err := i.getInfluxMetadata(metadata)
	if err != nil {
		return err
	}

	i.cfg = influxMeta
	if i.cfg.URL == "" {
		return ErrInfluxRequiredURL
	}

	if i.cfg.Token == "" {
		return ErrInfluxRequiredToken
	}

	if i.cfg.Org == "" {
		return ErrInfluxRequiredOrg
	}

	if i.cfg.Bucket == "" {
		return ErrInfluxRequiredBucket
	}

	log.Info("initialize timeseries.Influxdb", zap.String("url", i.cfg.URL))

	client := influxdb2.NewClient(i.cfg.URL, i.cfg.Token)
	i.client = client
	i.writeAPI = i.client.WriteAPIBlocking(i.cfg.Org, i.cfg.Bucket)
	i.queryAPI = i.client.QueryAPI(i.cfg.Org)

	return nil
}

// GetInfluxMetadata returns new Influx metadata.
func (i *Influx) getInfluxMetadata(metadata resource.TimeSeriesMetadata) (*InfluxConfig, error) {
	b, err := json.Marshal(metadata.Properties)
	if err != nil {
		return nil, errors.Wrap(err, "parse influx configurations")
	}

	var iMetadata InfluxConfig
	if err = json.Unmarshal(b, &iMetadata); err != nil {
		return nil, errors.Wrap(err, "parse influx configurations")
	}

	return &iMetadata, nil
}

// Invoke called on supported operations.
func (i *Influx) Write(ctx context.Context, req *WriteRequest) *Response {
	var points []string

	switch val := req.Data.(type) {
	case []string:
		points = val
	default:
		return &Response{Error: ErrInfluxInvalidParams}
	}

	// write the point.
	if err := i.writeAPI.WriteRecord(context.Background(), points...); err != nil {
		return &Response{Error: errors.Wrap(err, "write influxdb")}
	}

	i.client.Close()
	return &Response{Metadata: req.Metadata}
}

func (i Influx) Query(ctx context.Context, req QueryRequest) *Response {
	res, err := i.queryAPI.QueryRaw(ctx, req.ToRawQuery(), influxdb2.DefaultDialect())
	if err != nil {
		return &Response{Error: err}
	}
	return &Response{Raw: []byte(res)}
}

type FluxOption func(raw string) string

func FluxRangeOption(start, end string) FluxOption {
	if end == "" {
		end = "now()"
	}
	return func(raw string) string {
		raw += fmt.Sprintf("start: %s, stop: %s", start, end)
		return raw
	}
}

func FluxMeasurementFilterOption(operator, value string) FluxOption {
	return func(raw string) string {
		if strings.Contains(raw, "==") &&
			len(raw)-2 == strings.LastIndex(raw, "\"") {
			raw = strings.TrimSuffix(raw, "\n") + " " + operator + "\n"
		}
		raw += "r._measurement == " + fmt.Sprintf("%q\n", value)
		return raw
	}
}

func FluxFiledFilterOption(operator, value string) FluxOption {
	return func(raw string) string {
		if strings.Contains(raw, "==") &&
			len(raw)-2 == strings.LastIndex(raw, "\"") {
			raw = strings.TrimSuffix(raw, "\n") + " " + operator + "\n"
		}
		raw += "r._field == " + fmt.Sprintf("%q\n", value)
		return raw
	}
}

func FluxTagFilterOption(operator, value string) FluxOption {
	return func(raw string) string {
		if strings.Contains(raw, "==") &&
			len(raw)-2 == strings.LastIndex(raw, "\"") {
			raw = strings.TrimSuffix(raw, "\n") + " " + operator + "\n"
		}
		raw += "r.tag == " + fmt.Sprintf("%q\n", value)
		return raw
	}
}

func FluxGroupQueryOption(columns []string, mode ...string) FluxOption {
	for i := 0; i < len(columns); i++ {
		columns[i] = fmt.Sprintf("%q", columns[i])
	}
	return func(raw string) string {
		raw += fmt.Sprintf("columns:[%s]", strings.Join(columns, ","))
		if len(mode) != 0 {
			raw += fmt.Sprintf(", mode:%q", mode[0])
		}
		return raw
	}
}

func FluxSortOption(columns []string) FluxOption {
	for i := 0; i < len(columns); i++ {
		columns[i] = fmt.Sprintf("%q", columns[i])
	}
	return func(raw string) string {
		raw += fmt.Sprintf("columns: [%s]", strings.Join(columns, ","))
		return raw
	}
}

func FluxLimitOption(n int) FluxOption {
	return func(raw string) string {
		raw += fmt.Sprintf("n: %d", n)
		return raw
	}
}

type FluxQuery struct {
	FromBucket string
	RangeOpt   FluxOption
	FilterOpts []FluxOption
	GroupOpt   FluxOption
	SortOpt    FluxOption
	LimitOpt   FluxOption
	raw        string
}

func (q *FluxQuery) ToRawQuery() string {
	if q.raw == "" {
		q.raw = fmt.Sprintf("from(bucket: %q)\n", q.FromBucket)
	}
	raw := q.raw

	// add Range.
	if q.RangeOpt != nil {
		rangeStart := "|> range("
		rangeEnd := ")\n"
		raw += rangeStart
		raw = q.RangeOpt(raw) + rangeEnd
	}

	// add Filter.
	if len(q.FilterOpts) != 0 {
		filterStart := "|> filter(fn: (r) =>\n"
		filterEnd := ")\n"
		raw += filterStart
		for i := 0; i < len(q.FilterOpts); i++ {
			raw = q.FilterOpts[i](raw)
		}
		raw += filterEnd
	}

	// add Group.
	if q.GroupOpt != nil {
		groupStart := "|> group("
		groupEnd := ")\n"
		raw += groupStart
		raw = q.GroupOpt(raw) + groupEnd
	}

	// add Sort.
	if q.SortOpt != nil {
		sortStart := "|> sort("
		sortEnd := ")\n"
		raw += sortStart
		raw = q.SortOpt(raw) + sortEnd
	}

	// add Limit.
	if q.LimitOpt != nil {
		limitStart := "|> limit("
		limitEnd := ")"
		raw += limitStart
		raw = q.LimitOpt(raw) + limitEnd
	}

	q.raw = strings.TrimSuffix(raw, "\n")
	return q.raw
}
