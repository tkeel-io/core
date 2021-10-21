package service

import (
	"errors"
	"fmt"

	"github.com/tkeel-io/core/pkg/entities"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/logger"
)

var log = logger.NewLogger("core.api.service")

var (
	errTypeError      = errors.New("type error")
	errBodyMustBeJSON = errors.New("request body must be json")
)

func errResult(out *common.Content, err error) {
	if err != nil {
		out.Data = []byte(err.Error())
	}
}

const (
	entityFieldType   = "type"
	entityFieldID     = "id"
	entityFieldOwner  = "owner"
	entityFieldSource = "source"
)

func entityFieldRequired(fieldName string) error {
	return fmt.Errorf("entity field(%s) required", fieldName)
}

type Entity = entities.EntityBase
