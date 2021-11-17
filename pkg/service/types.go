package service

import (
	"github.com/tkeel-io/core/pkg/entities"

	"github.com/tkeel-io/core/pkg/logger"
)

var log = logger.NewLogger("core.api.service")

type Entity = entities.EntityBase

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
