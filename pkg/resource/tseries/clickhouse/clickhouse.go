package clickhouse

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	pb "github.com/tkeel-io/core/api/core/v1"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
)

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

type clickhouse struct {
	cfg          *Config
	msgQueue     chan *tseries.TSeriesData
	batchSize    int
	batchTimeout int
	conn         *sql.DB
}

func newClickhouse() tseries.TimeSerier {
	return &clickhouse{
		cfg:          &Config{},
		msgQueue:     make(chan *tseries.TSeriesData, 1000),
		batchSize:    100,
		batchTimeout: 1,
	}
}

func (c *clickhouse) getClickMetatadata(metadata resource.Metadata) (*Config, error) {
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

func (c *clickhouse) Init(meta resource.Metadata) error {
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
	go c.write()
	return nil
}

func (c *clickhouse) Write(ctx context.Context, req *tseries.TSeriesRequest) (*tseries.TSeriesResponse, error) {
	for _, item := range req.Data {
		c.msgQueue <- item
	}
	return &tseries.TSeriesResponse{}, nil
}

// 单列查，再拼接.
func (c *clickhouse) Query(ctx context.Context, req *pb.GetTSDataRequest) (*pb.GetTSDataResponse, error) {
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

func init() {
	tseries.Register("clickhouse", newClickhouse)
}

// 写入超时时间，和最大写入并发，每1秒或者100条写入一次.
func (c *clickhouse) write() {
	t := time.NewTimer(time.Second * time.Duration(c.batchTimeout))
	items := make([]*tseries.TSeriesData, 0, c.batchSize)
	for {
		select {
		case item := <-c.msgQueue:
			items = append(items, item)
			if len(items) >= c.batchSize {
				c.writeBatch(items)
				items = make([]*tseries.TSeriesData, 0, c.batchSize)
				t.Reset(time.Second * time.Duration(c.batchTimeout))
			}

		case <-t.C:
			if len(items) > 0 {
				c.writeBatch(items)
				items = make([]*tseries.TSeriesData, 0, c.batchSize)
			}
			t.Reset(time.Second * time.Duration(c.batchTimeout))
		}
	}
}

func (c *clickhouse) writeBatch(items []*tseries.TSeriesData) error {
	scope, err := c.conn.Begin()
	if err != nil {
		log.Error(err)
		return err
	}

	batch, err := scope.Prepare(fmt.Sprintf(ClickhouseSSQLTlp, c.cfg.Database, c.cfg.Table, "date, name, tags, value, timestamp"))
	if err != nil {
		return err
	}
	defer batch.Close()
	for _, item := range items {
		entityID, ok := item.Tags["id"]
		if !ok {
			continue
		}
		for k, v := range item.Fields {
			_, err := batch.Exec(
				time.UnixMilli(item.Timestamp/1e6),
				k,
				[]string{fmt.Sprintf("id=%s", entityID)},
				v,
				item.Timestamp/1e6,
			)
			if err != nil {
				log.Error(err)
				log.Info(item)
			}
		}
	}
	return scope.Commit()
}
