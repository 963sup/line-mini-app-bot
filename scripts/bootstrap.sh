#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source scripts/dev-env.sh
mode="${1:-core}"
case "$mode" in core|mcp|all) ;; *) echo "Usage: $0 [core|mcp|all]" >&2; exit 2;; esac
version=2026.11.3
case "$(uname -m)" in x86_64) arch=x64;; aarch64) arch=arm64;; *) exit 1;; esac
if [[ ! -x /workspace/tooling/bin/mise ]] || [[ "$(/workspace/tooling/bin/mise --version)" != "$version "* ]]; then
  stage="$(mktemp -d "$TMPDIR/mise-bootstrap.XXXXXX")"
  trap 'rm -rf "$stage"' EXIT
  artifact="mise-v$version-linux-$arch.tar.gz"
  base="https://github.com/jdx/mise/releases/download/v$version"
  # Release/CDN 429/5xx and timeouts are transient; bound retries and keep
  # checksum verification below authoritative before installing anything.
  download_options=(--fail --silent --show-error --location --retry 5
    --retry-max-time 180 --connect-timeout 15 --max-time 60)
  download_release_asset() {
    local name="$1" http_status asset_id
    if http_status="$(curl "${download_options[@]}" --write-out '%{http_code}' "$base/$name" -o "$stage/$name")"; then
      return
    fi
    # A persistent release-page gateway failure can be isolated from the API
    # route. Both routes serve the same pinned upstream asset, never a mirror.
    if [[ ! "$http_status" =~ ^(408|429|5[0-9][0-9])$ ]] || ! command -v gh >/dev/null 2>&1; then
      return 1
    fi
    echo "Release download returned HTTP $http_status; using the GitHub asset API for $name"
    asset_id="$(gh api "repos/jdx/mise/releases/tags/v$version" --jq ".assets[] | select(.name == \"$name\") | .id")" || return
    [[ "$asset_id" =~ ^[0-9]+$ ]] || return 1
    gh api -H 'Accept: application/octet-stream' "repos/jdx/mise/releases/assets/$asset_id" > "$stage/$name"
  }
  echo "Downloading pinned mise $version ($arch) and checksums"
  download_release_asset "$artifact"
  download_release_asset SHASUMS256.txt
  (cd "$stage"; awk -v file="$artifact" '{ name = $2; sub(/^\*/, "", name); sub(/^\.\//, "", name); if (name == file) print }' SHASUMS256.txt > selected.sha256; test -s selected.sha256; sha256sum -c selected.sha256)
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
