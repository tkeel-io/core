package runtime3

import "context"

type entity struct{}

func (s *entity) Handle(ctx context.Context, message interface{}) (*Result, error) {
	return &Result{State: []byte{}}, nil
}

func (s *entity) Raw() ([]byte, error) {
	return []byte{}, nil
}
