#!/usr/bin/env bash
set -euo pipefail
if [[ $# != 1 || "$1" == *...* ]]; then
  echo 'Usage: mise run profile ./path/to/one/package' >&2
  exit 2
fi
mkdir -p .artifacts/go
go test -run '^$' -bench . -benchmem -o .artifacts/go/profile.test -cpuprofile .artifacts/go/cpu.pprof -memprofile .artifacts/go/mem.pprof "$1"
