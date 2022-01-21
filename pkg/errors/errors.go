package errors

import "errors"

var (
	ErrInvalidJSONPath   = errors.New("Core.JSON.Path.Invalid")
	ErrInvalidProperties = errors.New("Core.Entity.Property.Invalid")
	ErrPropertyNotFound  = errors.New("Core.Entity.Property.NotFound")
	ErrInternal          = errors.New("Core.Internal")
	ErrEntityNotFound    = errors.New("Core.Entity.NotFound")
	ErrMapperNotFound    = errors.New("Core.Mapper.NotFound")
)
