#!/bin/bash
echo "Getting dependencies...."
go get -t
if [ $? -ne 0 ]; then
  echo "Failed to install dependencies!"
  exit 1
fi
echo "Running command"

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
