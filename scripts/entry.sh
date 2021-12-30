#!/bin/sh
echo "Running command "${1}""

return_empty() {
  text="${1}"
  echo "${text}" 1>&2
  if [ -n "${text}" ]; then
    echo 1
  else
    echo 0
  fi
}

imports() {
  exit $(return_empty "$(goimports -l .)")
}

formatting() {
  exit $(return_empty "$(gofmt -l .)")
}

vet() {
  exit $(return_empty "$(go vet .)")
}

ci_lint() {
  exit $(return_empty "$(golangci-lint run .)")
}

critic() {
  exit $(return_empty "$(gocritic check .)")
}

cycles() {
  exit $(return_empty "$(gocyclo -over 15 .)")
}

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
elif [ "${1}" = "imports" ]; then
  imports
elif [ "${1}" = "format" ]; then
  formatting
elif [ "${1}" = "vet" ]; then
  vet
elif [ "${1}" = "ci-lint" ]; then
  ci_lint
elif [ "${1}" = "critic" ]; then
  critic
elif [ "${1}" = "cycles" ]; then
  cycles
fi
