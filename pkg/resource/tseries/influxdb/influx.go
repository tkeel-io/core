package influxdb

import (
	"context"
	"encoding/json"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

// Influx allows writing to InfluxDB.
type Influx struct {
	cfg      *InfluxConfig
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
}

type InfluxConfig struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Org    string `json:"org"`
	Bucket string `json:"bucket"`
}

// NewInflux returns a new kafka binding instance.
func newInflux() tseries.TimeSerier {
	return &Influx{}
}

// Init does metadata parsing and connection establishment.
func (i *Influx) Init(metadata resource.Metadata) error {
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

	return nil
}

// GetInfluxMetadata returns new Influx metadata.
func (i *Influx) getInfluxMetadata(metadata resource.Metadata) (*InfluxConfig, error) {
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
func (i *Influx) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	var points []string

	switch val := req.Data.(type) {
	case []string:
		points = val
	default:
		return nil, ErrInfluxInvalidParams
	}

	// write the point.
	if err := i.writeAPI.WriteRecord(context.Background(), points...); err != nil {
		return nil, errors.Wrap(err, "write influxdb")
	}

	i.client.Close()
	return &tseries.TSeriesResponse{Metadata: req.Metadata}, nil
}

func init() {
	tseries.Register("influxdb", newInflux)
}
