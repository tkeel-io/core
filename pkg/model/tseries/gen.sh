#!/usr/bin/bash


# generare go from pb.
protoc --go_out=plugins=grpc:. tseries.proto