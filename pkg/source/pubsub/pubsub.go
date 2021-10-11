package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/source"
)

type PubSubMeta struct {
	name   string
	topics []string
	Pubsub string `json:"pubsub"`
	Topics string `json:"topics"`
}

type PubSubSource struct {
	name    string
	pubsub  string
	topics  []string
	service common.Service
	ctx     context.Context
}

func OpenSource(ctx context.Context, metadata source.Metadata, service common.Service) (source.ISource, error) {

	var (
		err  error
		meta *PubSubMeta
	)

	if meta, err = getMeta(metadata); err != nil {
		return nil, err
	}

	return &PubSubSource{
		ctx:     ctx,
		name:    meta.name,
		pubsub:  meta.Pubsub,
		topics:  meta.topics,
		service: service,
	}, nil
}

func (this *PubSubSource) String() string {
	return this.name
}

func (this *PubSubSource) StartReceiver(handler source.SourceHandler) error {
	for _, topic := range this.topics {
		if err := this.service.AddTopicEventHandler(
			&common.Subscription{
				PubsubName: this.pubsub,
				Topic:      topic,
			}, handler); err != nil {
			return err
		}
	}
	return nil
}

func (this *PubSubSource) Close() error {
	return errors.New("not implement.")
}

func getMeta(metadata source.Metadata) (*PubSubMeta, error) {
	b, err := json.Marshal(metadata.Properties)
	if err != nil {
		return nil, err
	}

	meta := PubSubMeta{}
	err = json.Unmarshal(b, &meta)
	if err != nil {
		return nil, err
	}

	meta.name = metadata.Name
	meta.topics = strings.Split(meta.Topics, ",")

	//meta.Name = metadata.Name

	//check name
	if "" == meta.Pubsub {
		return &meta, errors.New("field Name required.")
	} else if 0 == len(meta.topics) {
		return &meta, errors.New("field Topics required.")
	} else {
		return &meta, nil
	}
}

func init() {
	source.Register(&source.BaseSourceGenerator{SourceType: source.SourceTypePubSub, Generator: OpenSource})
}
