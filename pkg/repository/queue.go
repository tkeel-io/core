package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
)

func (r *repo) PutQueue(ctx context.Context, q *dao.Queue) error {
	return errors.Wrap(r.dao.PutQueue(ctx, q), "put queue repository")
}

func (r *repo) GetQueue(ctx context.Context, q *dao.Queue) (*dao.Queue, error) {
	queue, err := r.dao.GetQueue(ctx, q)
	return queue, errors.Wrap(err, "get queue repository")
}

func (r *repo) DelQueue(ctx context.Context, q *dao.Queue) error {
	return errors.Wrap(r.dao.DelQueue(ctx, q), "put queue repository")
}

func (r *repo) HasQueue(ctx context.Context, q *dao.Queue) (bool, error) {
	has, err := r.dao.HasQueue(ctx, q)
	return has, errors.Wrap(err, "exists queue repository")
}

func (r *repo) RangeQueue(ctx context.Context, rev int64, handler dao.QueueHandler) {
	r.dao.RangeQueue(ctx, rev, handler)
}

func (r *repo) WatchQueue(ctx context.Context, rev int64, handler dao.WatchQueueHandler) {
	go r.dao.WatchQueue(ctx, rev, handler)
}
