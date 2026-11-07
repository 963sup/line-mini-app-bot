#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source scripts/cloud-env.sh
mode="${1:-core}"
case "$mode" in core|mcp|all) ;; *) echo "Usage: $0 [core|mcp|all]" >&2; exit 2;; esac
version=2026.11.3
case "$(uname -m)" in x86_64) arch=x64;; aarch64) arch=arm64;; *) exit 1;; esac
if [[ ! -x /workspace/tooling/bin/mise ]] || [[ "$(/workspace/tooling/bin/mise --version)" != "$version "* ]]; then
  stage="$(mktemp -d "$TMPDIR/mise-bootstrap.XXXXXX")"
  trap 'rm -rf "$stage"' EXIT
  artifact="mise-v$version-linux-$arch.tar.gz"
  base="https://github.com/jdx/mise/releases/download/v$version"
  curl -fsSL "$base/$artifact" -o "$stage/$artifact"
  curl -fsSL "$base/SHASUMS256.txt" -o "$stage/SHASUMS256.txt"
  (cd "$stage"; awk -v file="$artifact" '$2 == file || $2 == "*"file { print }' SHASUMS256.txt > selected.sha256; test -s selected.sha256; sha256sum -c selected.sha256)
  tar -xzf "$stage/$artifact" -C "$stage"
  mkdir -p /workspace/tooling/bin
  install -m 0755 "$stage/mise/bin/mise" /workspace/tooling/bin/mise
fi
mise trust mise.toml
if [[ "$mode" == core || "$mode" == all ]]; then
  mise run install-core
  mise run doctor
fi
if [[ "$mode" == mcp || "$mode" == all ]]; then
  if [[ "$mode" == mcp ]]; then mise run install-core; fi
  mise run install-mcp
  mise run graphify-update
  mise run doctor-mcp
fi
