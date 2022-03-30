package discovery

import (
	"context"

	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func (d *Discovery) Register(ctx context.Context, node Service) error {
	var (
		err       error
		lease     clientv3.Lease
		leaseID   clientv3.LeaseID
		leaseResp *clientv3.LeaseGrantResponse
	)

	registerKey := node.Key()
	registerValue := node.Value()
	lease = clientv3.NewLease(d.discoveryEnd)
	if leaseResp, err = lease.Grant(ctx, d.HeartTime); err != nil {
		log.L().Error("grant lease", zap.Error(err))
		return errors.Wrap(err, "grant lease")
	}

	// register node.
	leaseID = leaseResp.ID
	_, err = d.discoveryEnd.Put(ctx, registerKey, registerValue, clientv3.WithLease(leaseID))
	if err != nil {
		log.L().Error("register service", zap.Error(err),
			zfield.Key(registerKey), zfield.Value(registerValue))
		return errors.Wrap(err, "register service")
	}

	log.L().Info("register service SUCCESS", zap.Error(err),
		zfield.Lease(int64(leaseID)), zfield.Key(registerKey), zfield.Value(registerValue))

	// keep lease alive.
	var leaseMessageCh <-chan *clientv3.LeaseKeepAliveResponse
	if leaseMessageCh, err = lease.KeepAlive(ctx, leaseID); nil != err {
		log.L().Error("keep lease alive", zap.Error(err),
			zfield.Lease(int64(leaseID)), zfield.Key(registerKey), zfield.Value(registerValue))
		return errors.Wrap(err, "keep lease alive")
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.L().Info("delete lease", zfield.Lease(int64(leaseID)))
				return
			case <-leaseMessageCh:
				// log.L().Debug("lease keepalive respose", zfield.Lease(int64(leaseID)), zfield.Cluster(leaseMsg.ClusterId),
				// 	zfield.Member(leaseMsg.MemberId), zfield.Revision(uint64(leaseMsg.Revision)), zfield.Term(int64(leaseMsg.RaftTerm)))
			}
		}
	}()

	return errors.Wrap(err, "keep lease alive")
}
