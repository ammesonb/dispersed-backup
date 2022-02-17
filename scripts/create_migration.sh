#!/bin/sh
if [ -z "${1}" ]; then
  echo "Must provide name for migration"
  exit 1
fi

migrate create -ext sql -dir mydb/migrations "${1}"
