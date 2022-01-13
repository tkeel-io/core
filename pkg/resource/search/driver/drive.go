package driver

import "context"

type Type string

type Engine interface {
	BuildIndex(ctx context.Context, index, content string) error
	Search(ctx context.Context, id string, offset, size int) (string, error)
	Delete(ctx context.Context, id string) error
}
