package influxdb

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
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
	queryAPI api.QueryAPI
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

	log.L().Info("initialize timeseries.Influxdb", zap.String("url", i.cfg.URL))

	client := influxdb2.NewClient(i.cfg.URL, i.cfg.Token)
	i.client = client
	i.writeAPI = i.client.WriteAPIBlocking(i.cfg.Org, i.cfg.Bucket)
	i.queryAPI = i.client.QueryAPI(i.cfg.Org)

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

func makeKVString(req map[string]string) string {
	ress := make([]string, 0)
	for k, v := range req {
		ress = append(ress, k+"="+v)
	}
	return strings.Join(ress, ",")
}

func makeKVSFloat(req map[string]float32) string {
	ress := make([]string, 0)
	for k, v := range req {
		ress = append(ress, fmt.Sprintf("%s=%f", k, v))
	}
	return strings.Join(ress, ",")
}

// Invoke called on supported operations.
func (i *Influx) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	points := make([]string, 0)
	for _, item := range req.Data {
		ss := fmt.Sprintf("%s,%s %s %d", item.Measurement, makeKVString(item.Tags), makeKVSFloat(item.Fields), item.Timestamp)
		points = append(points, ss)
	}
	// TODO: 时序 Client 有问题.
	if nil == i.writeAPI {
		return &tseries.TSeriesResponse{Metadata: req.Metadata}, nil
	}

	// write the point.
	if err := i.writeAPI.WriteRecord(ctx, points...); err != nil {
		return nil, errors.Wrap(err, "write influxdb")
	}
	return &tseries.TSeriesResponse{Metadata: req.Metadata}, nil
}

func (i *Influx) WriteData(req *pb.GetTSDataRequest) {
	points := make([]string, 1000)

	identifiers := strings.Split(req.Identifiers, ",")

	for i := 0; i < 1000; i++ {
		data := fmt.Sprintf("keel,id=%s ", req.Id)
		ident := make([]string, 0)
		for _, ideidentifier := range identifiers {
			value := float32(rand.Intn(100)) / 10.0 //nolint
			ident = append(ident, fmt.Sprintf("%s=%f", ideidentifier, value))
		}
		data = data + strings.Join(ident, ",") + fmt.Sprintf(" %d", (req.StartTime+(req.EndTime-req.StartTime)/int64(1000)*int64((i+1)))*1e9)
		//	points[i] = fmt.Sprintf("keel,id= avg=10.3,max=100.9 %d", time.Now().UnixNano()-(1000+int64(i))*1e9)
		points[i] = data
	}
	err := i.writeAPI.WriteRecord(context.Background(), points...)
	if err != nil {
		log.L().Error("write record", zap.Error(err))
	}
}

func (i *Influx) Query(ctx context.Context, req *pb.GetTSDataRequest) (*pb.GetTSDataResponse, error) {
	bucket := "core"
	measurement := "keel"
	startTime := req.StartTime
	endTime := req.EndTime
	entityID := req.Id
	offset := (req.PageNum - 1) * req.PageSize
	pageSize := req.PageSize

	resp := &pb.GetTSDataResponse{}

	queryString := `
	from(bucket: "%s")
    |> range(start: %d, stop: %d)
    |> filter(fn: (r) => r["_measurement"] == "%s")
    |> filter(fn: (r) => r["id"] == "%s")
    |> limit(n: %d, offset: %d)
	`

	querySS := fmt.Sprintf(queryString, bucket, startTime, endTime, measurement, entityID, pageSize, offset)
	identifiers := strings.Split(req.Identifiers, ",")
	identifiersItems := make([]string, 0)
	for _, identifier := range identifiers {
		identifiersItems = append(identifiersItems, fmt.Sprintf(`r._field == "%s"`, identifier))
	}
	if len(identifiersItems) > 0 {
		fieldString := strings.Join(identifiersItems, " or ")
		querySS = querySS + fmt.Sprintf(`|> filter(fn: (r) => %s)`, fieldString) + "\n"
	}

	resultPoints := make(map[time.Time]map[string]float32)

	result, err := i.queryAPI.Query(context.Background(), querySS)
	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				log.L().Info("Notice when group key has changed ",
					zap.String("table", result.TableMetadata().String()))
			}
			_, ok := resultPoints[result.Record().Time()]
			if !ok {
				resultPoints[result.Record().Time()] = make(map[string]float32)
			}

			floatVal, _ := result.Record().Value().(float64)
			resultPoints[result.Record().Time()][result.Record().Field()] = float32(floatVal)
		}
		// check for an error
		if result.Err() != nil {
			log.L().Error("quer influx database", zap.Error(err))
		}
	} else {
		log.L().Error("quer influx database", zap.Error(err))
	}

	for k, v := range resultPoints {
		resp.Items = append(resp.Items, &pb.TSResponse{
			Time:  k.UnixMilli(),
			Value: v,
		})
	}

	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].Time < resp.Items[j].Time
	})

	resp.Total = int32(len(resp.Items))
	//	if resp.Total == 0 {
	//		i.WriteData(req)
	//	}

	return resp, nil
}

func init() {
	tseries.Register("influxdb", newInflux)
}
