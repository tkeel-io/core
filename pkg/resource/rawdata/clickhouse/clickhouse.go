package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	ck "github.com/mailru/go-clickhouse"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type Clickhouse struct {
	option  *Option
	balance LoadBalance
}

func NewClickhouse() rawdata.RawDataService {
	return &Clickhouse{}
}

func (c *Clickhouse) parseOption(metadata resource.Metadata) (*Option, error) {
	opt := Option{}
	opt.DbName = metadata.Properties["database"].(string)
	opt.Fields = make(map[string]Field)
	opt.Table = metadata.Properties["table"].(string)
	items, ok := metadata.Properties["urls"].([]interface{})
	if !ok {
		return nil, errors.New("urls parse error")
	}
	for _, item := range items {
		opt.Urls = append(opt.Urls, item.(string))
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

const CLICKHOUSE_DB = `CREATE DATABASE IF NOT EXISTS %s`

const CLICKHOUSE_RAW_DATA = `CREATE TABLE %s.%s
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
			log.Error("open clickhouse", zap.Any("error", err))
			return err
		}
		if err = db.PingContext(context.Background()); err != nil {
			log.Error("ping clickhouse", zap.Any("error", err))
			return err
		}
		_, err = db.Exec(fmt.Sprintf(CLICKHOUSE_DB, opt.DbName))
		if err != nil {
			log.Warn(err.Error())
		}

		_, err = db.Exec(fmt.Sprintf(CLICKHOUSE_RAW_DATA, opt.DbName, opt.Table))
		if err != nil {
			log.Warn(err.Error())
		}
		if _, err = db.Query(fmt.Sprintf("desc %s.%s;", opt.DbName, opt.Table)); err != nil {
			log.Error("check chronus table", zap.Any("error", err))
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
func (c *Clickhouse) Write(ctx context.Context, req *rawdata.RawDataRequest) (err error) {
	//log.Info("chronus Insert ", logf.Any("messages", messages))
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

		//fmt.Println(string(message.Data()))
		log.L().Info("Invoke", zap.Any("messages", string(rawData.Bytes())))
		//jsonCtx := utils.NewJSONContext(string(message.Data()))

		data.fields = []string{"tag", "entity_id", "timestamp", "values", "path"}
		data.args = []interface{}{ck.Array(tags), rawData.EntityID, rawData.Timestamp.UnixMilli(), rawData.Values, rawData.Path}
		if len(data.fields) > 0 && len(data.fields) == len(data.args) {
			rows = append(rows, data)
		} else {
			log.L().Warn("rows is empty",
				zap.Any("args", data.args),
				zap.Any("fields", data.fields),
				zap.Any("option", c.option),
			)
		}
	}
	if len(rows) > 0 {
		preURL := c.genSql(rows[0])
		server := c.balance.Select([]*sqlx.DB{})
		if server == nil {
			return fmt.Errorf("get database failed, can't insert")
		}
		if tx, err = server.DB.BeginTx(ctx, nil); err != nil {
			log.L().Error("pre URL error",
				zap.String("preURL", preURL),
				zap.Any("row", rows[0]),
				zap.String("error", err.Error()))
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
				zap.String("preURL", preURL),
				zap.String("error", err.Error()))
			return err
		}
		for _, row := range rows {
			log.L().Debug("preURL",
				zap.Int64("ts", row.ts),
				zap.Any("args", row.args),
				zap.String("preURL", preURL))
			if _, err := stmt.Exec(row.args...); err != nil {
				log.L().Error("db Exec error",
					zap.String("preURL", preURL),
					zap.Any("args", row.args),
					zap.Any("fields", row.fields),
					zap.String("error", err.Error()))
				return err
			}
		}
		err = tx.Commit()
		if err != nil {
			row := rows[0]
			log.L().Error("tx Commit error",
				zap.Int64("ts", row.ts),
				zap.Any("args", row.args),
				zap.Any("fields", row.fields),
				zap.String("preURL", preURL),
				zap.String("error", err.Error()))
			return err
		}
		_ = stmt.Close()
	}
	return nil
}

func (c *Clickhouse) genSql(row *execNode) string {
	stmts := strings.Repeat("?,", len(row.fields))
	if len(stmts) > 0 {
		stmts = stmts[:len(stmts)-1]
	}
	return fmt.Sprintf(CLICKHOUSE_SSQL_TLP,
		c.option.DbName,
		c.option.Table,
		strings.Join(row.fields, ","),
		stmts)
}

func (c *Clickhouse) Query(ctx context.Context, req *pb.GetRawdataRequest) (resp *pb.GetRawdataResponse, err error) {
	var (
	//	tx *sql.Tx
	)

	querySql := "SELECT   `id`,`timestamp`, `entity_id`, `values`, `path`, `tag` FROM core.event_data1 where "
	countSql := "SELECT   count() FROM core.event_data1 where "

	querySql = querySql + fmt.Sprintf(" `timestamp` > FROM_UNIXTIME(%d) AND `timestamp` < FROM_UNIXTIME(%d)", req.StartTime, req.EndTime)
	countSql = countSql + fmt.Sprintf(" `timestamp` > FROM_UNIXTIME(%d) AND `timestamp` < FROM_UNIXTIME(%d)", req.StartTime, req.EndTime)

	querySql = querySql + fmt.Sprintf(" AND `entity_id`='%s'", req.EntityId)
	countSql = countSql + fmt.Sprintf(" AND `entity_id`='%s'", req.EntityId)

	querySql = querySql + fmt.Sprintf(" AND `path`='%s'", req.Path)
	countSql = countSql + fmt.Sprintf(" AND `path`='%s'", req.Path)

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
		fmt.Println(filterString)
		querySql = querySql + fmt.Sprintf(` AND hasAny(tag, [%s])`, filterString)
		countSql = countSql + fmt.Sprintf(` AND hasAny(tag, [%s])`, filterString)

	}

	if req.IsDescending {
		querySql = querySql + ` ORDER BY timestamp DESC`

	} else {

		querySql = querySql + ` ORDER BY timestamp ASC`
	}

	limit := req.PageSize
	offset := (req.PageNum - 1) * req.PageSize
	querySql = querySql + fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)

	fmt.Println(querySql)
	server := c.balance.Select([]*sqlx.DB{})

	countRes := make([]int64, 0)
	err = server.DB.SelectContext(context.Background(), &countRes, countSql)
	if err != nil {
		fmt.Println(err.Error())
	}

	queryRes := make([]*rawdata.RawData, 0)
	err = server.DB.SelectContext(context.Background(), &queryRes, querySql)
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, item := range queryRes {
		fmt.Println(item.Timestamp)
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

	return
}

func init() {
	rawdata.Register("clickhouse", NewClickhouse)
}
