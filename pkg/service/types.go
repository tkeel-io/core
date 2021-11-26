package service

import (
	"errors"

	"github.com/tkeel-io/core/pkg/statem"

	"github.com/tkeel-io/core/pkg/logger"
)

var log = logger.NewLogger("core.api.service")

var (
	ErrEntityMapperNil     = errors.New("mapper is nil")
	ErrEntityConfigInvalid = errors.New("entity config format invalid")
	ErrEntityInvalidParams = errors.New("invalid params")
	ErrEntityEmptyRequest  = errors.New("empty request")
)

type Entity = statem.Base

const (
	HeaderSource      = "Source"
	HeaderTopic       = "Topic"
	HeaderOwner       = "Owner"
	HeaderType        = "Type"
	HeaderMetadata    = "Metadata"
	HeaderContentType = "Content-Type"
	QueryType         = "type"

	Plugin = "plugin"
	User   = "user_id"
)
