package model

import (
	"context"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/google/uuid"
	"time"
)

const (
	EventTypeRaw          = "raw"
	EventTypeTS           = "time_series"
	EventTypeProperty     = "property"
	EventTypeRelationship = "relationship"
)

type KEvent struct {
	// ID of the event; must be non-empty and unique within the scope of the producer.
	// +required
	ID string `json:"id"`
	// Source - A URI describing the event producer.
	// +required
	Source string `json:"source"`
	// Type - The type of the occurrence which has happened.
	// +required
	Type string `json:"type"`
	// Topic
	// +required
	Topic string `json:"topic"`
	// User
	// +required
	User string `json:"user"`
	// DataContentType - A MIME (RFC2046) string describing the media type of `data`.
	// +optional
	DataContentType string `json:"data_content_type,omitempty"`
	// Time - A Timestamp when the event happened.
	// +optional
	Time time.Time `json:"time,omitempty"`
	// Data
	// +required
	Data []byte `json:"data"`
}

func NewKEventFromContext(ctx context.Context) (*KEvent, error) {
	kv := KEvent{}

	kv.Source = ctx.Value(service.HeaderSource).(string)
	if kv.Source == "" {
		return nil, SourceNilErr
	}

	kv.User = ctx.Value(service.HeaderUser).(string)
	if kv.User == "" {
		return nil, UserNilErr
	}

	kv.Topic = ctx.Value(service.HeaderTopic).(string)
	if kv.Topic == "" {
		return nil, TopicNilErr
	}

	kv.DataContentType = ctx.Value(service.HeaderContentType).(string)
	if kv.DataContentType == "" {
		return nil, DataContentTypeNilErr
	}
	kv.ID = uuid.New().String()
	kv.Time = time.Now()
	return &kv, nil
}
