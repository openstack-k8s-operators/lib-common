#!/bin/bash
set -ex

BASE_DIR="$(dirname $0)"
cd "${BASE_DIR}/.."

go test ./...
