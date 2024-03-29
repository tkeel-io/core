package clickhouse

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	sql "github.com/jmoiron/sqlx"

	"github.com/tkeel-io/core/pkg/resource/transport"

	jsoniter "github.com/json-iterator/go"

	pb "github.com/tkeel-io/core/api/core/v1"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const ClickhouseDBSQL = `CREATE DATABASE IF NOT EXISTS %s`

const ClickhouseTableSQL = `CREATE TABLE IF NOT EXISTS %s.%s 
(
	date Date DEFAULT toDate(0),
    name String,
    tags Array(String),
    value Float32,
    timestamp DateTime64(3, 'Asia/Shanghai'),
    updated DateTime64(3, 'Asia/Shanghai') DEFAULT now()
)
ENGINE = MergeTree
ORDER BY timestamp
SETTINGS index_granularity = 8192;
`

const (
	ClickhouseSSQLTlp = `INSERT INTO %s.%s (%s)`
	ClickHouseQuery   = `SELECT name, timestamp, value FROM %s.%s WHERE arrayExists(x -> x IN (%s), tags) AND `
)

type Config struct {
	Urls     []string `json:"urls"`
	Database string   `json:"database"`
	Table    string   `json:"table,omitempty"`
}

type Clickhouse struct {
	cfg  *Config
	conn *sql.DB
}

func newClickhouse() tseries.TimeSerier {
	return &Clickhouse{
		cfg: &Config{},
	}
}

func (c *Clickhouse) getClickMetatadata(metadata resource.Metadata) (*Config, error) {
	b, err := json.Marshal(metadata.Properties)
	if err != nil {
		return nil, errors.Wrap(err, "parse influx configurations")
	}

	var iMetadata Config
	if err = json.Unmarshal(b, &iMetadata); err != nil {
		return nil, errors.Wrap(err, "parse influx configurations")
	}

	return &iMetadata, nil
}

func (c *Clickhouse) Init(meta resource.Metadata) error {
	var err error
	c.cfg, err = c.getClickMetatadata(meta)
	if err != nil {
		return errors.Wrap(err, "clickhouse init error")
	}
	connectStr := fmt.Sprintf("%s?dial_timeout=1s&compress=true", c.cfg.Urls[0])
	conn, err := sql.Open("clickhouse", connectStr)
	if err != nil {
		log.Error("open clickhouse", logf.Any("error", err))
		return err
	}
	if err = conn.PingContext(context.Background()); err != nil {
		log.Error("ping clickhouse", logf.Any("error", err))
		return err
	}
	_, err = conn.Exec(fmt.Sprintf(ClickhouseDBSQL, c.cfg.Database))
	if err != nil {
		log.Warn(err.Error())
	}

	_, err = conn.Exec(fmt.Sprintf(ClickhouseTableSQL, c.cfg.Database, c.cfg.Table))
	if err != nil {
		log.Warn(err.Error())
	}
	if _, err = conn.Query(fmt.Sprintf("desc %s.%s;", c.cfg.Database, c.cfg.Table)); err != nil { //nolint
		log.Error("check chronus table", logf.Any("error", err))
		return err
	}
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Second)
	c.conn = conn
	//go c.write()
	return nil
}

func (c *Clickhouse) BatchWrite(ctx context.Context, args *[]interface{}) error {
	preURL := fmt.Sprintf(ClickhouseSSQLTlp, c.cfg.Database, c.cfg.Table, "date, name, tags, value, timestamp")
	if args != nil && len(*args) > 0 {
		return transport.BulkWrite(ctx, c.conn, preURL, args)
	}
	return errors.New("BatchWrite failed with：args == nil or len(*args) <= 0")
}

func (c *Clickhouse) BuildBulkData(req interface{}) (interface{}, error) {
	var argsVal = make([]interface{}, 0, 1)
	buildFn := func(req *tseries.TSeriesRequest, args *[]interface{}) {
		for _, item := range req.Data {
			entityID, ok := item.Tags["id"]
			if !ok {
				continue
			}
			timestamp := item.Timestamp / 1e6
			timeMilli := time.UnixMilli(timestamp)
			var builder strings.Builder
			builder.WriteString("id=")
			builder.WriteString(entityID)
			tagID := builder.String()
			for k, v := range item.Fields {
				*args = append(*args, []interface{}{timeMilli, k, []string{tagID}, v, timestamp})
			}
		}
	}

	if v, ok := req.(*tseries.TSeriesRequest); ok {
		buildFn(v, &argsVal)
		return argsVal, nil
	}
	return nil, errors.New("BuildBulkData error: invaild data")
}

func (c *Clickhouse) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	return &tseries.TSeriesResponse{}, nil
}

// 单列查，再拼接.
func (c *Clickhouse) Query(ctx context.Context, req *pb.GetTSDataRequest) (*pb.GetTSDataResponse, error) {
	resp := &pb.GetTSDataResponse{}
	tag := fmt.Sprintf(`'id=%s'`, req.GetId())
	querySQL := fmt.Sprintf(ClickHouseQuery, c.cfg.Database, c.cfg.Table, tag)
	querySQL += fmt.Sprintf(" `timestamp` > FROM_UNIXTIME(%d) AND `timestamp` < FROM_UNIXTIME(%d)", req.StartTime, req.EndTime)
	identifiers := strings.Split(req.Identifiers, ",")
	respData := make(map[time.Time]map[string]float32)
	for _, identifier := range identifiers {
		querySQL1 := querySQL + fmt.Sprintf(" AND name='%s' ", identifier)
		querySQL1 += `ORDER BY timestamp ASC`
		limit := req.PageSize
		offset := (req.PageNum - 1) * req.PageSize
		querySQL1 += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)

		rows, err := c.conn.Query(querySQL1)
		if err != nil {
			log.Error(err)
			continue
		}
		if rows.Err() != nil {
			log.Error(rows.Err())
			continue
		}
		defer rows.Close()
		for rows.Next() {
			var (
				identifier string
				t          time.Time
				value      float32
			)
			if err := rows.Scan(&identifier, &t, &value); err != nil {
				log.Error(err)
				continue
			}

			if _, ok := respData[t]; ok {
				respData[t][identifier] = value
			} else {
				respData[t] = make(map[string]float32)
				respData[t][identifier] = value
			}
		}
	}
	for k, v := range respData {
		resp.Items = append(resp.Items, &pb.TSResponse{
			Time:  k.UnixMilli(),
			Value: v,
		})
	}
	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].Time < resp.Items[j].Time
	})

	resp.Total = int32(len(resp.Items))
	resp.Total = int32(len(resp.Items))
	resp.PageNum = req.PageNum
	resp.PageSize = req.PageSize

	return resp, nil
}

func (c *Clickhouse) GetMetrics() (count, storage float64) {
	metricsSQL := fmt.Sprintf(`SELECT 
    	sum(rows) AS count,
    	sum(data_uncompressed_bytes) AS storage_uncompress,
    	sum(data_compressed_bytes) AS storage_compresss,
    	round((sum(data_compressed_bytes) / sum(data_uncompressed_bytes)) * 100, 0) AS compress_rate
	FROM system.parts WHERE (active = 1) AND (database = '%s') AND(table = '%s')
	GROUP BY partition
	ORDER BY partition ASC`, c.cfg.Database, c.cfg.Table)
	rows, err := c.conn.Query(metricsSQL)
	if err != nil {
		log.Error(err)
		return
	}
	if rows.Err() != nil {
		log.Error(rows.Err())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			storageUncompress float64
			storageCompresss  float64
			compressRate      float64
		)
		if err := rows.Scan(&count, &storageUncompress, &storageCompresss, &compressRate); err != nil {
			log.Error(err)
			continue
		}
		return count, storageCompresss
	}
	return count, storage
}
