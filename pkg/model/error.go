package model

import "github.com/pkg/errors"

var ErrTopicNil = errors.New("topic is nil")
var ErrSourceNil = errors.New("source is nil")
var ErrUserNil = errors.New("user is nil")
var ErrDataContentTypeNil = errors.New("data_content_type is nil")
var ErrEventType = errors.New("event type error")
