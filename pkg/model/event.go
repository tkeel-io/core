package model

import (
	"context"
	"time"

	"github.com/tkeel-io/core/pkg/service"

	"github.com/google/uuid"
)

const (
	EventTypeRaw          = "raw"
	EventTypeTS           = "time_series"
	EventTypeProperty     = "property"
	EventTypeRelationship = "relationship"
)

type KEvent struct {
	// ID of the event; must be non-empty and unique within the scope of the producer.
	// +required.
	ID string `json:"id"`
	// Source - A URI describing the event producer.
	// +required.
	Source string `json:"source"`
	// Type - The type of the occurrence which has happened.
	// +required.
	Type string `json:"type"`
	// Topic
	// +required.
	Topic string `json:"topic"`
	// User
	// +required.
	User string `json:"user"`
	// DataContentType - A MIME (RFC2046) string describing the media type of `data`.
	// +optional.
	DataContentType string `json:"data_content_type,omitempty"`
	// Time - A Timestamp when the event happened.
	// +optional.
	Time time.Time `json:"time,omitempty"`
	// Data
	// +required.
	Data []byte `json:"data"`
}

func newKEvent() KEvent {
	return KEvent{
		ID:              "",
		Source:          "",
		Type:            "",
		Topic:           "",
		User:            "",
		DataContentType: "",
		Time:            time.Now(),
		Data:            nil,
	}
}

func NewKEventFromContext(ctx context.Context) (*KEvent, error) {
	kv := newKEvent()

	var ok bool
	if kv.Source, ok = ctx.Value(service.HeaderSource).(string); !ok || kv.Source == "" {
		return nil, ErrSourceNil
	}

	if kv.User, ok = ctx.Value(service.HeaderUser).(string); !ok || kv.User == "" {
		return nil, ErrUserNil
	}

	if kv.Topic, ok = ctx.Value(service.HeaderTopic).(string); !ok || kv.Topic == "" {
		return nil, ErrTopicNil
	}

	if kv.DataContentType, ok = ctx.Value(service.HeaderContentType).(string); !ok || kv.DataContentType == "" {
		return nil, ErrDataContentTypeNil
	}

	kv.ID = uuid.New().String()
	kv.Time = time.Now()

	return &kv, nil
}
