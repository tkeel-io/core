package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

const (
	SubscriptionPrefix = "/core/v1/subscription"
)

type ListSubscriptionReq struct {
	Owner    string
	EntityID string
}

var _ dao.Resource = (*Subscription)(nil)

type Subscription struct {
	// Subscription identifier.
	ID string
	// Subscription name.
	Name string
	// Subscription owner.
	Owner string
	// Subscription.
	Subscription string
	// description.
	Description string
}

func NewSubscription(owner, ID, name, subscription, desc string) *Subscription {
	return &Subscription{
		ID:          ID,
		Name:        name,
		Owner:       owner,
		Subscription:      subscription,
		Description: desc,
	}
}

func ListSubscriptionPrefix(Owner, EntityID string) string {
	keyString := fmt.Sprintf("%s/%s",
		SubscriptionPrefix, Owner)
	return keyString
}

func (s *Subscription) EncodeKey() ([]byte, error) {
	if s.Owner == ""{
		return nil, errors.Errorf("Subscription Owner is empty")
	}
	if s.ID == ""{
		return nil, errors.Errorf("Subscription ID is empty")
	}

	keyString := fmt.Sprintf("%s/%s/%s",
		SubscriptionPrefix, s.Owner, s.ID)
	return []byte(keyString), nil
}

func (s *Subscription) Encode() ([]byte, error) {
	bytes, err := json.Marshal(s)
	return bytes, errors.Wrap(err, "encode Subscription")
}

func (s *Subscription) Decode(bytes []byte) error {
	err := json.Unmarshal(bytes, s)
	return errors.Wrap(err, "decode Subscription")
}

func (r *repo) PutSubscription(ctx context.Context, expr Subscription) error {
	err := r.dao.PutResource(ctx, &expr)
	return errors.Wrap(err, "put expression repository")
}

func (r *repo) GetSubscription(ctx context.Context, expr Subscription) (Subscription, error) {
	_, err := r.dao.GetResource(ctx, &expr)
	return expr, errors.Wrap(err, "get expression repository")
}

func (r *repo) DelSubscription(ctx context.Context, expr Subscription) error {
	err := r.dao.DelResource(ctx, &expr)
	return errors.Wrap(err, "del expression repository")
}

func (r *repo) HasSubscription(ctx context.Context, expr Subscription) (bool, error) {
	has, err := r.dao.HasResource(ctx, &expr)
	return has, errors.Wrap(err, "exists expression repository")
}

func (r *repo) ListSubscription(ctx context.Context, rev int64, req *ListSubscriptionReq) ([]*Subscription, error) {
	// construct prefix.
	prefix := ListSubscriptionPrefix(req.EntityID, req.Owner)
	ress, err := r.dao.ListResource(ctx, rev, prefix,
		func(raw []byte) (dao.Resource, error) {
			var res Subscription // escape.
			err := res.Decode(raw)
			return &res, errors.Wrap(err, "decode expression")
		})

	var exprs []*Subscription
	for index := range ress {
		if expr, ok := ress[index].(*Subscription); ok {
			exprs = append(exprs, expr)
			continue
		}
		// panic.
	}
	return exprs, errors.Wrap(err, "list expression repository")
}

func (r *repo) RangeSubscription(ctx context.Context, rev int64, handler RangeSubscriptionFunc) {
	r.dao.RangeResource(ctx, rev, SubscriptionPrefix, func(kvs []*mvccpb.KeyValue) {
		var exprs []*Subscription
		for index := range kvs {
			var expr Subscription
			err := expr.Decode(kvs[index].Value)
			if nil != err {
				log.L().Error("")
				continue
			}
			exprs = append(exprs, &expr)
		}
		handler(exprs)
	})
}

func (r *repo) WatchSubscription(ctx context.Context, rev int64, handler WatchSubscriptionFunc) {
	r.dao.WatchResource(ctx, rev, SubscriptionPrefix, func(et dao.EnventType, kv *mvccpb.KeyValue) {
		var expr Subscription
		err := expr.Decode(kv.Value)
		if nil != err {
			log.L().Error("")
		}
		handler(et, expr)
	})
}

type RangeSubscriptionFunc func([]*Subscription)
type WatchSubscriptionFunc func(dao.EnventType, Subscription)
