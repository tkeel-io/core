package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	// Subscription owner.
	Owner string

	Mode              string `protobuf:"bytes,1,opt,name=mode,proto3" json:"mode,omitempty"`
	Source            string `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	Filter            string `protobuf:"bytes,3,opt,name=filter,proto3" json:"filter,omitempty"`
	Target            string `protobuf:"bytes,4,opt,name=target,proto3" json:"target,omitempty"`
	Topic             string `protobuf:"bytes,5,opt,name=topic,proto3" json:"topic,omitempty"`
	PubsubName        string `protobuf:"bytes,6,opt,name=pubsub_name,json=pubsubName,proto3" json:"pubsub_name,omitempty"`
	Source2           string

	SourceEntityPaths []string
	SourceEntityID    string
}

func NewSubscription(ID, Owner, Mode, Source, Filter, Target, Topic, PubsubName string) *Subscription {
	return &Subscription{
	}
}

func ListSubscriptionPrefix(Owner, EntityID string) string {
	keyString := fmt.Sprintf("%s/%s",
		SubscriptionPrefix, Owner)
	return keyString
}

func (s *Subscription) EncodeKey() ([]byte, error) {
	if s.Owner == "" {
		return nil, errors.Errorf("Subscription Owner is empty")
	}
	if s.ID == "" {
		return nil, errors.Errorf("Subscription ID is empty")
	}

	keyString := fmt.Sprintf("%s/%s/%s/%s",
		SubscriptionPrefix, s.Owner, s.ID, s.SourceEntityID)
	return []byte(keyString), nil
}

func (s *Subscription) Encode() ([]byte, error) {
	bytes, err := json.Marshal(s)
	return bytes, errors.Wrap(err, "encode Subscription")
}

func (s *Subscription) Decode(key, bytes []byte) error {
	if bytes != nil {
		err := json.Unmarshal(bytes, s)
		return errors.Wrap(err, "decode Subscription")
	} else {
		///core/v1/subscription/admin/sub-1234/device123
		keys := strings.Split(string(key), "/")
		if len(keys) != 7 {
			return errors.Errorf("error:decode Subscription from key[%s]", string(key))
		}
		s.Owner = keys[4]
		s.ID = keys[5]
		s.SourceEntityID = keys[6]
		return nil
	}
}

func (r *repo) PutSubscription(ctx context.Context, expr *Subscription) error {
	err := r.dao.PutResource(ctx, expr)
	return errors.Wrap(err, "put expression repository")
}

func (r *repo) GetSubscription(ctx context.Context, expr *Subscription) (*Subscription, error) {
	_, err := r.dao.GetResource(ctx, expr)
	return expr, errors.Wrap(err, "get expression repository")
}

func (r *repo) DelSubscription(ctx context.Context, expr *Subscription) error {
	err := r.dao.DelResource(ctx, expr)
	return errors.Wrap(err, "del expression repository")
}

func (r *repo) HasSubscription(ctx context.Context, expr *Subscription) (bool, error) {
	has, err := r.dao.HasResource(ctx, expr)
	return has, errors.Wrap(err, "exists expression repository")
}

func (r *repo) ListSubscription(ctx context.Context, rev int64, req *ListSubscriptionReq) ([]*Subscription, error) {
	// construct prefix.
	prefix := ListSubscriptionPrefix(req.EntityID, req.Owner)
	ress, err := r.dao.ListResource(ctx, rev, prefix,
		func(key, raw []byte) (dao.Resource, error) {
			var res Subscription // escape.
			err := res.Decode(key, raw)
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
			err := expr.Decode(kvs[index].Key, kvs[index].Value)
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
		var expr = &Subscription{}
		err := expr.Decode(kv.Key, kv.Value)
		if nil != err {
			log.L().Error("")
		}
		handler(et, expr)
	})
}

type RangeSubscriptionFunc func([]*Subscription)
type WatchSubscriptionFunc func(dao.EnventType, *Subscription)
