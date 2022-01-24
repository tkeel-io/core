package influxdb

import (
	"context"
	"os"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

// Influx allows writing to InfluxDB.
type Influx struct {
	id       string
	cfg      *InfluxConfig
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
}

type InfluxConfig struct {
	URL    string `mapstructure:"url"`
	Token  string `mapstructure:"token"`
	Org    string `mapstructure:"org"`
	Bucket string `mapstructure:"bucket"`
}

// Invoke called on supported operations.
func (i *Influx) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	switch points := req.Data.(type) {
	case []string:
		// write the point.
		if err := i.writeAPI.WriteRecord(context.Background(), points...); err != nil {
			return nil, errors.Wrap(err, "write influxdb")
		}
	default:
		return nil, ErrInfluxInvalidParams
	}

	i.client.Close()
	return &tseries.TSeriesResponse{Metadata: req.Metadata}, nil
}

func init() {
	zfield.SuccessStatusEvent(os.Stdout, "Register Resource<TSDB.influxdb> successful")
	tseries.Register("influxdb", func(properties map[string]interface{}) (tseries.TimeSerier, error) {
		var err error
		var influxMeta InfluxConfig
		if err = mapstructure.Decode(properties, &influxMeta); err != nil {
			return nil, errors.Wrap(err, "mapstructure decode")
		}

		if influxMeta.URL == "" {
			return nil, ErrInfluxRequiredURL
		}

		if influxMeta.Token == "" {
			return nil, ErrInfluxRequiredToken
		}

		if influxMeta.Org == "" {
			return nil, ErrInfluxRequiredOrg
		}

		if influxMeta.Bucket == "" {
			return nil, ErrInfluxRequiredBucket
		}

		log.Info("initialize timeseries.Influxdb", zap.String("url", influxMeta.URL))

		client := influxdb2.NewClient(influxMeta.URL, influxMeta.Token)
		writeAPI := client.WriteAPIBlocking(influxMeta.Org, influxMeta.Bucket)

		id := util.UUID()
		log.Info("create pubsub.noop instance", zfield.ID(id))

		return &Influx{
			id:       id,
			cfg:      &influxMeta,
			client:   client,
			writeAPI: writeAPI,
		}, nil
	})
}
