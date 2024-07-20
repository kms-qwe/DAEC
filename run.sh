#!/bin/sh
go run ./cmd/storage/main.go --config=./config/local.yaml --reset=true
go run ./cmd/auth/main.go --config=./config/local.yaml &
go run ./cmd/orch/main.go --config=./config/local.yaml &
go run ./cmd/agent/main.go --config=./config/local.yaml &