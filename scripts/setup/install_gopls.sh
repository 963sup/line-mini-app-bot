#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
version=$(awk '$1 == "go" {print $2; exit}' "$root/backend/go.mod")
bin="$HOME/.local/bin"
sdk="$HOME/.local/go-$version"
export PATH="$bin:$PATH"
if (( $# > 1 )) || [[ $# == 1 && $1 != --update ]]; then
  echo "Usage: $0 [--update]" >&2
  exit 2
fi

tmp=$(mktemp -d)
cleanup() {
  chmod -R u+w "$tmp"
  rm -rf -- "$tmp"
}
trap cleanup EXIT
mkdir -p "$bin"
if [[ $(go version 2>/dev/null || true) != "go version go$version "* ]]; then
  if [[ ! -x "$sdk/bin/go" ]]; then
    case "$(uname -sm)" in
      'Linux x86_64') platform=linux-amd64 ;;
      'Linux aarch64') platform=linux-arm64 ;;
      *) echo "Install Go $version for this platform first." >&2; exit 1 ;;
    esac
    archive="go$version.$platform.tar.gz"
    digest=$(curl -fsSL 'https://go.dev/dl/?mode=json&include=all' |
      jq -er --arg file "$archive" '.[] | .files[] | select(.filename == $file) | .sha256')
    curl -fsSL "https://dl.google.com/go/$archive" -o "$tmp/sdk.tgz"
    echo "$digest  $tmp/sdk.tgz" | sha256sum -c -
    tar -xzf "$tmp/sdk.tgz" -C "$tmp"
    test ! -e "$sdk"
    mv -T "$tmp/go" "$sdk"
  fi
  ln -s "$sdk/bin/go" "$bin/go"
fi

if [[ ${1:-} == --update ]] || ! gopls help mcp >/dev/null 2>&1; then
  GOPATH="$tmp/gopath" GOCACHE="$tmp/cache" GOTOOLCHAIN=local GOBIN="$tmp/bin" \
    go install golang.org/x/tools/gopls@latest
  install -m 755 "$tmp/bin/gopls" "$bin/gopls"
  hash -r
fi
gopls version
gopls help mcp >/dev/null
