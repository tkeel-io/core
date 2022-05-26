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
	option  *Option
	balance LoadBalance
}

func NewClickhouse() rawdata.Service {
	return &Clickhouse{}
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
	return nil
}

func (c *Clickhouse) Write(ctx context.Context, req *rawdata.Request) (err error) {
	// log.Info("chronus Insert ", logf.Any("messages", messages)).
	var (
		tx *sql.Tx
	)
	rows := make([]*execNode, 0)
	tags := make([]string, 0)
	for k, v := range req.Metadata {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}
	for _, rawData := range req.Data {
		data := new(execNode)

		// fmt.Println(string(message.Data()))
		log.L().Info("Invoke", logf.Any("messages", string(rawData.Bytes())))
		// jsonCtx := utils.NewJSONContext(string(message.Data()))

		data.fields = []string{"tag", "entity_id", "timestamp", "values", "path"}
		data.args = []interface{}{tags, rawData.EntityID, rawData.Timestamp.UnixMilli(), rawData.Values, rawData.Path}
		if len(data.fields) > 0 && len(data.fields) == len(data.args) {
			rows = append(rows, data)
		} else {
			log.L().Warn("rows is empty",
				logf.Any("args", data.args),
				logf.Any("fields", data.fields),
				logf.Any("option", c.option),
			)
		}
	}
	if len(rows) > 0 {
		preURL := c.genSQL(rows[0])
		server := c.balance.Select([]*sqlx.DB{})
		if server == nil {
			return fmt.Errorf("get database failed, can't insert")
		}
		if tx, err = server.DB.BeginTx(ctx, nil); err != nil {
			log.L().Error("pre URL error",
				logf.String("preURL", preURL),
				logf.Any("row", rows[0]),
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
		for _, row := range rows {
			log.L().Debug("preURL",
				logf.Int64("ts", row.ts),
				logf.Any("args", row.args),
				logf.String("preURL", preURL))
			if _, err = stmt.Exec(row.args...); err != nil {
				log.L().Error("db Exec error",
					logf.String("preURL", preURL),
					logf.Any("args", row.args),
					logf.Any("fields", row.fields),
					logf.String("error", err.Error()))
				return err
			}
		}
		err = tx.Commit()
		if err != nil {
			row := rows[0]
			log.L().Error("tx Commit error",
				logf.Int64("ts", row.ts),
				logf.Any("args", row.args),
				logf.Any("fields", row.fields),
				logf.String("preURL", preURL),
				logf.String("error", err.Error()))
			return err
		}
	}
	return nil
}

func (c *Clickhouse) genSQL(row *execNode) string {
	stmts := strings.Repeat("?,", len(row.fields))
	if len(stmts) > 0 {
		stmts = stmts[:len(stmts)-1]
	}
	return fmt.Sprintf(ClickhouseSSQLTlp,
		c.option.DbName,
		c.option.Table,
		strings.Join(row.fields, ","),
		stmts)
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

func (c *Clickhouse) GetMetrics() (count, storage float64) {
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
		return count, storageCompresss
	}
	return count, storage
}

func init() {
	rawdata.Register("clickhouse", NewClickhouse)
}
