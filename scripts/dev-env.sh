# Source this file; values are owned exclusively by mise.toml.
# The base image must provide Python 3.11+; mise Python is installed later.
if ! command -v python3 >/dev/null 2>&1; then
  printf "%s\n" "Bootstrap requires Python 3.11+ on PATH before mise is installed." >&2
  return 1
fi
_dev_env_exports="$(python3 "$(dirname "${BASH_SOURCE[0]}")/render-env.py")" || return
eval "$_dev_env_exports"
unset _dev_env_exports
mkdir -p "$XDG_CACHE_HOME" "$XDG_CONFIG_HOME" "$XDG_DATA_HOME" "$XDG_STATE_HOME" "$XDG_RUNTIME_DIR" "$TMPDIR" "$GOTMPDIR" "$GOMODCACHE" "$GOCACHE" "$(dirname "$GOENV")" "$(dirname "$npm_config_userconfig")" "$MISE_STATE_DIR" "$CLOUDSDK_CONFIG" "$DOCKER_CONFIG"
chmod 700 "$XDG_RUNTIME_DIR" "$TMPDIR"
case ":$PATH:" in
  *:/workspace/tooling/bin:*) ;;
  *) export PATH="/workspace/tooling/bin:$PATH" ;;
esac
