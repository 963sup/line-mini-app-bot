# Source this file in each shell; keep host HOME/authentication unchanged.
export XDG_CACHE_HOME=/workspace/tooling/cache
export XDG_CONFIG_HOME=/workspace/tooling/config
export XDG_DATA_HOME=/workspace/tooling/data
export XDG_STATE_HOME=/workspace/tooling/state
export XDG_RUNTIME_DIR=/workspace/tooling/runtime
export TMPDIR=/workspace/tooling/tmp
export MISE_DATA_DIR=/workspace/tooling/mise-data
export MISE_CACHE_DIR="$XDG_CACHE_HOME/mise"
export MISE_STATE_DIR="$XDG_STATE_HOME/mise"
export MISE_CONFIG_DIR="$XDG_CONFIG_HOME/mise"
export MISE_SELF_UPDATE_AVAILABLE=false
export GOPATH=/workspace/tooling/go
export GOCACHE="$XDG_CACHE_HOME/go-build"
export GOMODCACHE="$GOPATH/pkg/mod"
export GOENV="$XDG_CONFIG_HOME/go/env"
export GOTMPDIR="$TMPDIR/go"
export GOTOOLCHAIN=local
export GOFLAGS=-mod=readonly
export npm_config_cache="$XDG_CACHE_HOME/npm"
export npm_config_prefix=/workspace/tooling/npm
export npm_config_userconfig="$XDG_CONFIG_HOME/npm/npmrc"
export UV_CACHE_DIR="$XDG_CACHE_HOME/uv"
export UV_TOOL_DIR="$XDG_DATA_HOME/uv/tools"
export UV_TOOL_BIN_DIR=/workspace/tooling/bin
export UV_PYTHON_INSTALL_DIR="$XDG_DATA_HOME/uv/python"
export PIP_CACHE_DIR="$XDG_CACHE_HOME/pip"
export PIPX_HOME="$XDG_DATA_HOME/pipx"
export PIPX_BIN_DIR=/workspace/tooling/bin
export PYTHONPYCACHEPREFIX="$XDG_CACHE_HOME/python"
export PLAYWRIGHT_BROWSERS_PATH="$XDG_CACHE_HOME/ms-playwright"
export HISTFILE="$XDG_STATE_HOME/bash_history"
mkdir -p "$XDG_CACHE_HOME" "$XDG_CONFIG_HOME" "$XDG_DATA_HOME" "$XDG_STATE_HOME" "$XDG_RUNTIME_DIR" "$TMPDIR" "$GOTMPDIR" "$GOMODCACHE" "$GOCACHE" "$XDG_CONFIG_HOME/go" "$XDG_CONFIG_HOME/npm" "$MISE_STATE_DIR"
chmod 700 "$XDG_RUNTIME_DIR" "$TMPDIR"
case ":$PATH:" in
  *:/workspace/tooling/bin:*) ;;
  *) export PATH="/workspace/tooling/bin:$PATH" ;;
esac
# Preserve an explicitly selected cloud identity/configuration.
export CLOUDSDK_CONFIG="${CLOUDSDK_CONFIG:-/workspace/tooling/config/gcloud}"
export DOCKER_CONFIG="${DOCKER_CONFIG:-/workspace/tooling/config/docker}"
mkdir -p "$CLOUDSDK_CONFIG" "$DOCKER_CONFIG"
