#!/bin/bash

# Find all proto files recursively
PROTO_FILES=$(find proto -name "*.proto")

if [ -z "$PROTO_FILES" ]; then
  echo "No proto files found"
  exit 1
fi

protoc \
  --proto_path=. \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  $PROTO_FILES