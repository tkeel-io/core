package transport

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
)

type ClickHouseCli interface {
	BatchWrite(ctx context.Context, reqs []interface{}) error
}

type BulkWriteFunc func(ctx context.Context, db *sqlx.DB, query string, args *[][]interface{}) error

func BulkWrite(ctx context.Context, db *sqlx.DB, query string, args *[][]interface{}) error {
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
	for _, val := range *args {
		if _, err = stmt.Exec(val...); err != nil {
			return err
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
			err = cli.BatchWrite(context.Background(), messages)
			if err != nil {
				log.Println(err)
			}
			return err
		})
	return sinkTransport, err
}
