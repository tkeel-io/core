package errors

import "errors"

var (
	ErrInvalidJSONPath   = errors.New("invalid JSONPath")
	ErrInvalidProperties = errors.New("statem invalid properties")
	ErrPropertyNotFound  = errors.New("property not found")
	ErrInternal          = errors.New("internel error")
)
