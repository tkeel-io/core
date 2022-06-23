package transport

import (
	"context"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

type ClickHouseCli interface {
	BatchWrite(ctx context.Context, args *[]interface{}) error
	BuildBulkData(m interface{}) (interface{}, error)
}

func BulkWrite(ctx context.Context, db *sqlx.DB, query string, args *[]interface{}) error {
	var (
		err error
		tx  *sql.Tx
	)

	if tx, err = db.BeginTx(ctx, nil); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, arg := range *args {
		rows, ok := arg.([]interface{})
		if !ok {
			continue
		}
		for _, row := range rows {
			if rowData, ok := row.([]interface{}); ok {
				if _, err = stmt.Exec(rowData...); err != nil {
					return err
				}
			}
		}
	}
	err = tx.Commit()
	return err
}

func NewClickHouseTransport(ctx context.Context, cli ClickHouseCli) (Transport, error) {
	sinkTransport, err := NewSinkTransport(
		ctx,
		"clickhouse",
		func(messages []interface{}) (err error) {
			err = cli.BatchWrite(context.Background(), &messages)
			if err != nil {
				log.Println(err)
			}
			return err
		},
		cli.BuildBulkData)
	return sinkTransport, err
}
