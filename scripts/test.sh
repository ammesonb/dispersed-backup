#!/bin/sh
echo "Running command "${1}""

if [ "${1}" = "test" ]; then
  go test -v ./...
elif [ "${1}" = "lint" ]; then
  echo "Imports"
  goimports -l .
  echo "Formatting"
  go fmt
  echo "Vetting"
  go vet
  echo "Linting"
  golint
  echo "CI Lint"
  golangci-lint run
  echo "Critique"
  gocritic check
  echo "Cycles"
  gocyclo -over 15 .
fi
