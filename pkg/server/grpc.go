package server

import (
	"github.com/tkeel-io/kit/transport/grpc"
)

// NewHTTPServer new a GRPC server.
func NewGRPCServer(addr string) *grpc.Server {
	s := grpc.NewServer(addr)
	return s
}
