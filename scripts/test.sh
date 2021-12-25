#!/bin/bash
go get -t

if [ "${1}" = "ci" ]; then
  go test -v ./...
elif [ "${1}" = "lint" ]; then
  goimports
  go fmt
  go vet
  golint
  golangci-lint
  gocritic check
  gocyclo -over 15
fi
