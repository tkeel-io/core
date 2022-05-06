package dapr

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
)

type daprMetadata struct {
	TopicName  string `json:"topic_name" mapstructure:"topic_name"`
	PubsubName string `json:"pubsub_name" mapstructure:"pubsub_name"`
}

func New(id, urlText string) (pubsub.Pubsub, error) {
	daprMeta, err := parseURL(urlText)
	if nil != err {
		log.Error("parse url configuration", logf.Error(err), logf.URL(urlText))
		return nil, errors.Wrap(err, "parse url")
	}

	pid := util.UUID("pubsub.dapr")
	log.L().Info("create pubsub.dapr instance", logf.ID(pid))

	return &daprPubsub{
		id:         pid,
		topicName:  daprMeta.TopicName,
		pubsubName: daprMeta.PubsubName,
	}, nil
}

type daprPubsub struct {
	id         string
	topicName  string
	pubsubName string
}

func (d *daprPubsub) ID() string {
	return d.id
}

func (d *daprPubsub) Send(ctx context.Context, event v1.Event) error {
	panic("never used")
}

func (d *daprPubsub) Received(ctx context.Context, handler pubsub.EventHandler) error {
	log.L().Debug("pubsub.dapr start receive message", logf.ID(d.id))
	Register(&Consumer{id: d.id, handler: handler})
	return errors.Wrap(nil, "register message handler")
}

func (d *daprPubsub) Commit(v interface{}) error {
	return nil
}

func (d *daprPubsub) Close() error {
	log.L().Debug("pubsub.dapr close", logf.ID(d.id))
	Unregister(&Consumer{id: d.id})
	return errors.Wrap(nil, "unregister message handler")
}

func init() {
	log.SuccessStatusEvent(os.Stdout, "Register Resource<pubsub.dapr> successful")
	pubsub.Register("dapr", func(id string, urlText string) (pubsub.Pubsub, error) {
		pubsubIns, err := New(id, urlText)
		return pubsubIns, errors.Wrap(err, "new pubsub.dapr instance")
	})
}

// dapr://hosts/pubsub_name/topic
func parseURL(urlText string) (*daprMetadata, error) {
	urlIns, err := url.Parse(urlText)
	if nil != err {
		return nil, errors.Wrap(err, "parse configuration from url")
	}

	segs := strings.Split(urlIns.Path, "/")
	if len(segs) != 3 {
		return nil, xerrors.ErrInvalidParam
	}

	return &daprMetadata{
		TopicName:  segs[2],
		PubsubName: segs[1],
	}, nil
}
