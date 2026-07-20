#!/usr/bin/env bash
# Install the exact tool versions required by scripts/ci/ci.sh.

set -euo pipefail

go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0
go install github.com/loov/goda@v0.9.4
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
