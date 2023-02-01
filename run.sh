#!/bin/sh

go build -o ./build/main ./cmd/main.go

./build/main \
  --endpoint "ws://localhost:1337/ws" \
  --sender "0x..." \
  --senderPrivateKeyHex "..." \
  --recipient "0x..."
