#!/usr/bin/env bash

set -euo pipefail

taskfile="Taskfile.yml"

go_version="$(sed -n "s/^  GO_VERSION: '\(.*\)'/\1/p" "$taskfile" | head -n1)"
golangci_lint_version="$(sed -n "s/^  GOLANGCI_LINT_VERSION: '\(.*\)'/\1/p" "$taskfile" | head -n1)"
modules="$(sed -n "s/^  MODULES: '\(.*\)'/\1/p" "$taskfile" | head -n1)"
integration_modules="$(sed -n "s/^  INTEGRATION_MODULES: '\(.*\)'/\1/p" "$taskfile" | head -n1)"

{
  echo "GO_VERSION=$go_version"
  echo "GOLANGCI_LINT_VERSION=$golangci_lint_version"
  echo "MODULES=$modules"
  echo "INTEGRATION_MODULES=$integration_modules"
} >> "$GITHUB_OUTPUT"
