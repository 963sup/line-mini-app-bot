# Source this file; values are owned exclusively by mise.toml.
_cloud_env_exports="$(python3 "$(dirname "${BASH_SOURCE[0]}")/load-env.py")" || return
eval "$_cloud_env_exports"
unset _cloud_env_exports
mkdir -p "$XDG_CACHE_HOME" "$XDG_CONFIG_HOME" "$XDG_DATA_HOME" "$XDG_STATE_HOME" "$XDG_RUNTIME_DIR" "$TMPDIR" "$GOTMPDIR" "$GOMODCACHE" "$GOCACHE" "$(dirname "$GOENV")" "$(dirname "$npm_config_userconfig")" "$MISE_STATE_DIR" "$CLOUDSDK_CONFIG" "$DOCKER_CONFIG"
chmod 700 "$XDG_RUNTIME_DIR" "$TMPDIR"
case ":$PATH:" in
  *:/workspace/tooling/bin:*) ;;
  *) export PATH="/workspace/tooling/bin:$PATH" ;;
esac
