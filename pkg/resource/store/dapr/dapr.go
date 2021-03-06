package dapr

import (
	"context"
	"os"
	"strings"

	"github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/core/pkg/resource/transport"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/store"
	"github.com/tkeel-io/core/pkg/util/dapr"

	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type daprMetadata struct {
	StoreName string `mapstructure:"store_name"`
}

type daprBulkStore struct {
	daprStore
	bulkTransport transport.Transport
}

func (d *daprBulkStore) Set(ctx context.Context, key string, data []byte) error {
	return d.bulkTransport.Send(ctx, &client.SetStateItem{Key: key, Value: data})
}

func (d *daprBulkStore) Flush(ctx context.Context) error {
	return d.bulkTransport.Flush(ctx)
}

type daprStore struct {
	id        string
	storeName string
}

// Get returns state.
func (d *daprStore) Get(ctx context.Context, key string) (*store.StateItem, error) {
	var conn dapr.Client
	if conn = dapr.Get().Select(); nil == conn {
		log.L().Error("nil connection", logf.Key(key),
			logf.String("store_name", d.storeName), logf.ID(d.id))
		return nil, errors.Wrap(xerrors.ErrConnectionNil, "dapr send")
	}

	item, err := conn.GetState(ctx, d.storeName, key)
	if nil != err {
		return nil, errors.Wrap(err, "dapr store get")
	}

	if len(item.Value) == 0 {
		return nil, xerrors.ErrEntityNotFound
	}

	return &store.StateItem{
		Key:      item.Key,
		Etag:     item.Etag,
		Value:    item.Value,
		Metadata: item.Metadata,
	}, nil
}

// Set saves the raw data into store using default state options.
func (d *daprStore) Set(ctx context.Context, key string, data []byte) error {
	var conn dapr.Client
	if conn = dapr.Get().Select(); nil == conn {
		log.L().Error("nil connection", logf.Key(key),
			logf.String("store_name", d.storeName),
			logf.ID(d.id), logf.String("data", string(data)))
		return errors.Wrap(xerrors.ErrConnectionNil, "dapr send")
	}
	return errors.Wrap(conn.SaveState(ctx, d.storeName, key, data), "dapr store set")
}

func (d *daprStore) BatchWrite(ctx context.Context, args *[]interface{}) error {
	var conn dapr.Client
	stateMap := make(map[string]*client.SetStateItem)
	for _, val := range *args {
		// need to order by update-time
		if item, ok := val.(*client.SetStateItem); ok {
			stateMap[item.Key] = item
		} else {
			return errors.Wrap(errors.New("invalid data"), "daprStore BatchWrite args error")
		}
	}
	BulkStateItems := make([]*client.SetStateItem, 0, len(stateMap))
	for _, v := range stateMap {
		BulkStateItems = append(BulkStateItems, v)
	}
	if conn = dapr.Get().Select(); nil == conn {
		var buf strings.Builder
		for _, k := range BulkStateItems {
			if _, err := buf.WriteString(k.Key); err != nil {
				log.L().Error(err.Error(),
					logf.String("daprStore BatchWrite buf.WriteString(k.Key) err", k.Key))
			}
		}
		log.L().Error("nil connection", logf.Key(buf.String()),
			logf.String("store_name", d.storeName),
			logf.ID(d.id))
		return errors.Wrap(xerrors.ErrConnectionNil, "dapr send")
	}
	return errors.Wrap(conn.SaveBulkState(ctx, d.storeName, BulkStateItems...), "dapr store set")
}

func (d *daprStore) BuildBulkData(m interface{}) (interface{}, error) {
	//TODO
	return m, nil
}

func (d *daprStore) Del(ctx context.Context, key string) error {
	var conn dapr.Client
	if conn = dapr.Get().Select(); nil == conn {
		log.L().Error("nil connection", logf.Key(key),
			logf.String("store_name", d.storeName), logf.ID(d.id))
		return errors.Wrap(xerrors.ErrConnectionNil, "dapr send")
	}
	return errors.Wrap(conn.DeleteState(ctx, d.storeName, key), "dapr store del")
}

func init() {
	log.SuccessStatusEvent(os.Stdout, "Register Resource<state.dapr> successful")
	store.Register("dapr", func(properties map[string]interface{}) (store.Store, error) {
		var daprMeta daprMetadata
		if err := mapstructure.Decode(properties, &daprMeta); nil != err {
			return nil, errors.Wrap(err, "decode store.dapr configuration")
		}

		id := util.UUID("sdapr")
		log.L().Info("create store.dapr instance", logf.ID(id))
		s := daprStore{
			id:        id,
			storeName: daprMeta.StoreName,
		}
		bulkTransport, err := transport.NewDaprStateTransport(context.Background(), &s)
		if err != nil {
			return nil, err
		}
		return &daprBulkStore{
			daprStore:     s,
			bulkTransport: bulkTransport,
		}, nil
	})
}
