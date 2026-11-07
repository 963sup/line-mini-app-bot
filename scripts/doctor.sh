#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source scripts/cloud-env.sh
mise exec -- go version
for name in GOPATH GOCACHE GOMODCACHE GOENV GOTMPDIR; do
  value="$(mise exec -- go env "$name")"
  case "$value" in /workspace/*) ;; *) echo "FAIL: $name outside workspace: $value"; exit 1;; esac
done
for directory in "$GOCACHE" "$GOMODCACHE" "$GOTMPDIR" "$XDG_CONFIG_HOME" "$CLOUDSDK_CONFIG" "$DOCKER_CONFIG"; do
  test -w "$directory" || { echo "FAIL: not writable: $directory"; exit 1; }
done
checkdir="$(mktemp -d "$TMPDIR/go-doctor.XXXXXX")"
trap 'rm -rf "$checkdir"' EXIT
printf 'module doctor.local/check\n\ngo 1.27.3\n' > "$checkdir/go.mod"
printf 'package check\nfunc Sum(a,b int) int { return a+b }\n' > "$checkdir/check.go"
printf 'package check\nimport "testing"\nfunc TestSum(t *testing.T) { if Sum(2,3)!=5 { t.Fatal("sum") } }\n' > "$checkdir/check_test.go"
mise exec -- bash -c 'set -e; cd "$1"; go test -race ./...; go vet ./...; go build -buildvcs=false ./...' -- "$checkdir"
echo "PASS: Go compiler, race detector, vet and writable directories"
if [[ -f go.mod ]]; then
  echo "Application module exists; run mise run check separately"
else
  echo "NOT RUN: application checks (no go.mod)"
fi
for tool in gopls gcloud docker java; do
  if mise exec -- sh -c 'command -v "$1" >/dev/null' -- "$tool"; then
    echo "PRESENT: $tool (presence only; not service or authentication readiness)"
  else
    echo "OPTIONAL/MISSING: $tool"
  fi
done
