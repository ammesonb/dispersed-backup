#!/bin/bash
echo "Getting dependencies...."
echo $(which go)
echo $GOROOT
echo $GOPATH
echo $(ls /usr/src/go/)
go get -t
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
