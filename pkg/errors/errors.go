package errors

import "errors"

var (
	ErrInvalidJSONPath          = errors.New("Core.JSON.Path.Invalid")
	ErrInvalidProperties        = errors.New("Core.Entity.Property.Invalid")
	ErrPropertyNotFound         = errors.New("Core.Entity.Property.NotFound")
	ErrInternal                 = errors.New("Core.Internal")
	ErrEntityNotFound           = errors.New("Core.Entity.NotFound")
	ErrEntityAleadyExists       = errors.New("Core.Entity.Already.Exists")
	ErrInvalidEntityParams      = errors.New("Core.Entity.Params.Invalid")
	ErrRuntimeNotExists         = errors.New("Core.Runtime.NotExists")
	ErrMapperNotFound           = errors.New("Core.Mapper.NotFound")
	ErrQueueNotFound            = errors.New("Core.Queue.NotFound")
	ErrNodeNotExist             = errors.New("Core.Cluster.Node.NotExist")
	ErrInvalidQueueType         = errors.New("Core.Queue.Type.Invalid")
	ErrInvalidQueueConsumerType = errors.New("Core.Queue.Consumer.Type.Invalid")
	ErrInvalidMessageType       = errors.New("Core.Message.Type.Invalid")
	ErrInvalidMessageField      = errors.New("Core.Message.Field.Invalid")
	ErrInvalidSubscriptionMode  = errors.New("Core.Message.Mode.Invalid")
	ErrInvalidPropertyConfig    = errors.New("Core.Entity.Property.Config.Invalid")
	ErrInvalidHTTPRequest       = errors.New("Core.Transport.Http.Request.Invalid")
	ErrInvalidHTTPInited        = errors.New("Core.Transport.Http.Inited")
	ErrTemplateNotFound         = errors.New("Core.Template.NotFound")
	ErrEntityPropertyIDEmpty    = errors.New("Core.Entity.PropertyID.Empty")
	ErrInvalidRequest           = errors.New("Core.Request.Invalid")
	ErrEntityConfigInvalid      = errors.New("invalid entity configurations")
	ErrJSONPatchReservedOp      = errors.New("invalid json patch operator")
	ErrInvalidNodeType          = errors.New("undefine node type")
	ErrEmptyParam               = errors.New("empty params")
	ErrPatchNotFound            = errors.New("patch not found")
	ErrPatchPathInvalid         = errors.New("invalid patch path")
	ErrPatchPathLack            = errors.New("patch path lack")
	ErrPatchPathRoot            = errors.New("patch path lack root")
	ErrPatchTypeInvalid         = errors.New("patch config type invalid")
	ErrServerNotReady           = errors.New("Core.Service.NotReady")
	ErrConnectionNil            = errors.New("Core.Resource.Connection.Nil")
	ErrInvalidParam             = errors.New("Core.Params.Invalid")
	ErrExpressionNotFound       = errors.New("Core.Expression.NotFound")

	// Resource errors.
	ErrResourceNotFound = errors.New("Core.Resource.NotFound")
)

func New(code string) error {
	return errors.New(code)
}
