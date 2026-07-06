#!/bin/bash

protoc \
  --go_out=. \
  --go-grpc_out=. \
  proto/**/*.proto

go get google.golang.org/protobuf