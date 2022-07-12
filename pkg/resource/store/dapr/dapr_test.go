package dapr

import (
	"context"
	"testing"

	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/resource/transport"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

func TestDaprStateTransport(t *testing.T) {
	id := util.UUID("sdapr")
	log.L().Info("create store.dapr instance", logf.ID(id))
	s := daprStore{
		id:        id,
		storeName: "core-state",
	}
	bulkTransport, err := transport.NewDaprStateTransport(context.Background(), &s)
	if err != nil {
		t.Error(err)
	}
	bulkStore := daprBulkStore{
		daprStore:     s,
		bulkTransport: bulkTransport,
	}
	err = bulkStore.Set(context.Background(), "test1", []byte("test2"))
	if err != nil {
		t.Error(err)
	}
	//ch := make(chan os.Signal, 1)
	//signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	//<-ch
}
