package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	logf "github.com/tkeel-io/core/pkg/logfield"

	"github.com/jmoiron/sqlx"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/kit/log"
)

const ClickhouseDBSQL = `CREATE DATABASE IF NOT EXISTS %s`

const ClickhouseTableSQL = `CREATE TABLE IF NOT EXISTS %s.%s 
(
    id UUID DEFAULT generateUUIDv4(),
    entity_id String,
    path String,
    timestamp DateTime64(3, 'Asia/Shanghai'),
    tag Array(String),
    values String
)
ENGINE = MergeTree
ORDER BY timestamp
SETTINGS index_granularity = 8192;
`

type Clickhouse struct {
	option       *Option
	balance      LoadBalance
	msgQueue     chan *rawdata.Request
	batchSize    int
	batchTimeout int
}

func NewClickhouse() rawdata.Service {
	return &Clickhouse{
		msgQueue:     make(chan *rawdata.Request, 3000),
		batchSize:    100,
		batchTimeout: 1,
	}
}

func (c *Clickhouse) parseOption(metadata resource.Metadata) (*Option, error) {
	opt := Option{}
	var ok bool
	if opt.DbName, ok = metadata.Properties["database"].(string); !ok {
		return nil, errors.New("config error")
	}
	opt.Fields = make(map[string]Field)
	if opt.Table, ok = metadata.Properties["table"].(string); !ok {
		return nil, errors.New("config error")
	}
	items, ok := metadata.Properties["urls"].([]interface{})
	if !ok {
		return nil, errors.New("urls parse error")
	}
	for _, item := range items {
		itemStr, ok := item.(string)
		if !ok {
			log.Warn("url config is not string")
			continue
		}
		opt.Urls = append(opt.Urls, itemStr)
	}

	if opt.Fields == nil {
		return nil, errors.New("field not found")
	}

	for key, field := range opt.Fields {
		if key == "" {
			return nil, errors.New("field name is empty")
		}
		if field.Type == "" {
			return nil, fmt.Errorf("field(%s) types is empty", key)
		}
		if field.Value == "" {
			return nil, fmt.Errorf("field(%s) types is empty", key)
		}
	}
	return &opt, nil
}

func (c *Clickhouse) Init(metadata resource.Metadata) error {
	opt, err := c.parseOption(metadata)
	if err != nil {
		return err
	}
	servers := make([]*Server, len(opt.Urls))
	for k, v := range opt.Urls {
		log.Info("clickhouse init " + v)
		db, err := sqlx.Open("clickhouse", v)
		if err != nil {
			log.Error("open clickhouse", logf.Any("error", err))
			return err
		}
		if err = db.PingContext(context.Background()); err != nil {
			log.Error("ping clickhouse", logf.Any("error", err))
			return err
		}
		_, err = db.Exec(fmt.Sprintf(ClickhouseDBSQL, opt.DbName))
		if err != nil {
			log.Warn(err.Error())
		}

		_, err = db.Exec(fmt.Sprintf(ClickhouseTableSQL, opt.DbName, opt.Table))
		if err != nil {
			log.Warn(err.Error())
		}
		if _, err = db.Query(fmt.Sprintf("desc %s.%s;", opt.DbName, opt.Table)); err != nil { //nolint
			log.Error("check chronus table", logf.Any("error", err))
			return err
		}
		db.SetConnMaxLifetime(30 * time.Second)
		db.SetMaxOpenConns(5)
		servers[k] = &Server{db, v, 1}
	}
	c.option = opt

	c.balance = NewLoadBalanceRandom(servers)
	go c.write()
	return nil
}

func (c *Clickhouse) Write(ctx context.Context, req *rawdata.Request) (err error) {
	c.msgQueue <- req
	return
}

// 写入超时时间，和最大写入并发，每1秒或者100条写入一次.
func (c *Clickhouse) write() {
	t := time.NewTimer(time.Second * time.Duration(c.batchTimeout))
	items := make([]*rawdata.Request, 0, c.batchSize)
	for {
		select {
		case item := <-c.msgQueue:
			items = append(items, item)
			if len(items) >= c.batchSize {
				c.writeBatch(items)
				items = make([]*rawdata.Request, 0, c.batchSize)
				t.Reset(time.Second * time.Duration(c.batchTimeout))
			}

		case <-t.C:
			if len(items) > 0 {
				c.writeBatch(items)
				items = make([]*rawdata.Request, 0, c.batchSize)
			}
			t.Reset(time.Second * time.Duration(c.batchTimeout))
		}
	}
}

func (c *Clickhouse) writeBatch(items []*rawdata.Request) (err error) {
	// log.Info("chronus Insert ", logf.Any("messages", messages)).
	var (
		tx *sql.Tx
	)
	server := c.balance.Select([]*sqlx.DB{})
	if server == nil {
		return fmt.Errorf("get database failed, can't insert")
	}
	fields := []string{"tag", "entity_id", "timestamp", "values", "path"}
	preURL := fmt.Sprintf(ClickhouseSSQLTlp, c.option.DbName, c.option.Table, strings.Join(fields, ","))
	if tx, err = server.DB.BeginTx(context.Background(), nil); err != nil {
		log.L().Error("pre URL error",
			logf.String("preURL", preURL),
			logf.Any("row", fields),
			logf.String("error", err.Error()))
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare(preURL)
	if err != nil {
		log.L().Error("pre URL error",
			logf.String("preURL", preURL),
			logf.String("error", err.Error()))
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		tags := make([]string, 0)
		for k, v := range item.Metadata {
			tags = append(tags, fmt.Sprintf("%s=%s", k, v))
		}
		for _, rawData := range item.Data {
			log.L().Debug("Invoke", logf.Any("messages", string(rawData.Bytes())))
			args := []interface{}{tags, rawData.EntityID, rawData.Timestamp.UnixMilli(), rawData.Values, rawData.Path}
			if _, err = stmt.Exec(args...); err != nil {
				log.L().Error("db Exec error",
					logf.String("preURL", preURL),
					logf.Any("args", args),
					logf.String("error", err.Error()))
			}
		}
	}
	return tx.Commit()
}

func (c *Clickhouse) Query(ctx context.Context, req *pb.GetRawdataRequest) (resp *pb.GetRawdataResponse, err error) {
	querySQL := fmt.Sprintf("SELECT   `id`,`timestamp`, `entity_id`, `values`, `path`, `tag` FROM %s.%s where ", c.option.DbName, c.option.Table)
	countSQL := fmt.Sprintf("SELECT   count() FROM %s.%s where ", c.option.DbName, c.option.Table)

	querySQL += fmt.Sprintf(" `timestamp` > FROM_UNIXTIME(%d) AND `timestamp` < FROM_UNIXTIME(%d)", req.StartTime, req.EndTime)
	countSQL += fmt.Sprintf(" `timestamp` > FROM_UNIXTIME(%d) AND `timestamp` < FROM_UNIXTIME(%d)", req.StartTime, req.EndTime)

	querySQL += fmt.Sprintf(" AND `entity_id`='%s'", req.EntityId)
	countSQL += fmt.Sprintf(" AND `entity_id`='%s'", req.EntityId)

	querySQL += fmt.Sprintf(" AND `path`='%s'", req.Path)
	countSQL += fmt.Sprintf(" AND `path`='%s'", req.Path)

	filters := make([]string, 0)
	if req.Filters != nil {
		for k, v := range req.Filters {
			items := strings.Split(v, ",")
			for _, item := range items {
				filters = append(filters, fmt.Sprintf("'%s=%s'", k, item))
			}
		}
	}
	filterString := strings.Join(filters, ",")
	if filterString != "" {
		querySQL += fmt.Sprintf(` AND hasAny(tag, [%s])`, filterString)
		countSQL += fmt.Sprintf(` AND hasAny(tag, [%s])`, filterString)
	}

	if req.IsDescending {
		querySQL += ` ORDER BY timestamp DESC`
	} else {
		querySQL += ` ORDER BY timestamp ASC`
	}

	limit := req.PageSize
	offset := (req.PageNum - 1) * req.PageSize
	querySQL += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)

	server := c.balance.Select([]*sqlx.DB{})

	countRes := make([]int64, 0)
	err = server.DB.SelectContext(context.Background(), &countRes, countSQL)
	if err != nil {
		log.Error(err.Error())
		return nil, pb.ErrClickhouse()
	}

	queryRes := make([]*rawdata.RawData, 0)
	err = server.DB.SelectContext(context.Background(), &queryRes, querySQL)
	if err != nil {
		log.Error(err.Error())
		return nil, pb.ErrClickhouse()
	}
	resp = &pb.GetRawdataResponse{
		Total:    0,
		PageNum:  0,
		PageSize: 0,
		Items:    []*pb.RawdataResponse{},
	}
	for _, item := range queryRes {
		resp.Items = append(resp.Items, &pb.RawdataResponse{
			Timestamp: item.Timestamp.UnixMilli(),
			Id:        item.ID,
			EntityId:  item.EntityID,
			Path:      item.Path,
			Values:    item.Values,
		})
	}
	if len(countRes) == 1 {
		resp.Total = int32(countRes[0])
	}
	resp.PageNum = req.PageNum
	resp.PageSize = req.PageSize
	return resp, err
}

func (c *Clickhouse) getSystemSpace() (total, used float64) {
	metricsSQL := "SELECT free_space, total_space FROM system.disks"
	server := c.balance.Select([]*sqlx.DB{})
	rows, err := server.DB.Query(metricsSQL)
	if err != nil {
		log.Error(err)
		return
	}
	if rows.Err() != nil {
		log.Error(rows.Err())
		return
	}
	defer rows.Close()
	var totalAll, freeAll float64
	for rows.Next() {
		var freeSpace, totalSpace float64

		if err := rows.Scan(&freeSpace, &totalSpace); err != nil {
			log.Error(err)
			continue
		} else {
			totalAll += totalSpace
			freeAll += freeSpace
		}
	}
	return totalAll, totalAll - freeAll
}

func (c *Clickhouse) GetMetrics() (count, storage, total, used float64) {
	total, used = c.getSystemSpace()
	metricsSQL := fmt.Sprintf(`SELECT 
    	sum(rows) AS count,
    	sum(data_uncompressed_bytes) AS storage_uncompress,
    	sum(data_compressed_bytes) AS storage_compresss,
    	round((sum(data_compressed_bytes) / sum(data_uncompressed_bytes)) * 100, 0) AS compress_rate
	FROM system.parts WHERE (active = 1) AND (database = '%s') AND(table = '%s')
	GROUP BY partition
	ORDER BY partition ASC`, c.option.DbName, c.option.Table)

	server := c.balance.Select([]*sqlx.DB{})
	rows, err := server.DB.Query(metricsSQL)
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
		return count, storageCompresss, total, used
	}
	return count, storage, total, used
}

func init() {
	rawdata.Register("clickhouse", NewClickhouse)
}
