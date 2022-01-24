package errors

import "errors"

var (
	ErrInvalidJSONPath          = errors.New("Core.JSON.Path.Invalid")
	ErrInvalidProperties        = errors.New("Core.Entity.Property.Invalid")
	ErrPropertyNotFound         = errors.New("Core.Entity.Property.NotFound")
	ErrInternal                 = errors.New("Core.Internal")
	ErrEntityNotFound           = errors.New("Core.Entity.NotFound")
	ErrMapperNotFound           = errors.New("Core.Mapper.NotFound")
	ErrQueueNotFound            = errors.New("Core.Queue.NotFound")
	ErrNodeNotExist             = errors.New("Core.Cluster.Node.NotExist")
	ErrInvalidQueueType         = errors.New("Core.Queue.Type.Invalid")
	ErrInvalidQueueConsumerType = errors.New("Core.Queue.Consumer.Type.Invalid")
	ErrMessageTypeInvalid       = errors.New("Core.Message.Type.Invalid")
)
