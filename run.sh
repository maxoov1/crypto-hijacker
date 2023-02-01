#!/bin/sh

export HIJACKER_ENDPOINT=""

export HIJACKER_SENDER=""
export HIJACKER_SENDER_PRIVATE_KEY=""

export HIJACKER_RECIPIENT=""

set -x

go build -o ./build/main ./cmd/main.go
./build/main
