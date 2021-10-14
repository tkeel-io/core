package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/source"
)

type Meta struct {
	name   string
	topics []string
	Pubsub string `json:"pubsub"`
	Topics string `json:"topics"`
}

type Source struct {
	name    string
	pubsub  string
	topics  []string
	service common.Service
	ctx     context.Context
}

func OpenSource(ctx context.Context, metadata source.Metadata, service common.Service) (source.ISource, error) {
	var (
		err  error
		meta *Meta
	)

	if meta, err = acquireMeta(metadata); err != nil {
		return nil, err
	}

	return &Source{
		ctx:     ctx,
		name:    meta.name,
		pubsub:  meta.Pubsub,
		topics:  meta.topics,
		service: service,
	}, nil
}

func (s *Source) String() string {
	return s.name
}

func (s *Source) StartReceiver(handler source.Handler) error {
	for _, topic := range s.topics {
		if err := s.service.AddTopicEventHandler(
			&common.Subscription{
				PubsubName: s.pubsub,
				Topic:      topic,
			}, handler); err != nil {
			return errors.Unwrap(err)
		}
	}
	return nil
}

func (s *Source) Close() error {
	return errors.New("not implement")
}

func acquireMeta(metadata source.Metadata) (*Meta, error) {
	b, err := json.Marshal(metadata.Properties)
	if err != nil {
		return nil, errors.Unwrap(err)
	}

	meta := Meta{}
	err = json.Unmarshal(b, &meta)
	if err != nil {
		return nil, errors.Unwrap(err)
	}

	meta.name = metadata.Name
	meta.topics = strings.Split(meta.Topics, ",")

	// meta.Name = metadata.Name

	// check name
	if meta.Pubsub == "" {
		return &meta, errors.New("field Name required")
	} else if 0 == len(meta.topics) {
		return &meta, errors.New("field Topics required")
	} else {
		return &meta, nil
	}
}

func init() {
	source.Register(&source.BaseSourceGenerator{SourceType: source.PubSub, Generator: OpenSource})
}
