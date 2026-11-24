#!/usr/bin/env bash
# Linux x86_64. Requires bash, curl, tar, sha256sum, flock and the project's Go version.
# Run from any directory. Arguments after the script name are passed to go test.
set -euo pipefail

backend=$(cd "$(dirname "$0")/../../backend" && pwd)
cache=${FIRESTORE_TOOL_CACHE:-${XDG_CACHE_HOME:-$HOME/.cache}/line-mini-app-bot/firestore}
port=${FIRESTORE_TEST_PORT:-18080}
[[ $(uname -sm) == 'Linux x86_64' ]] || { echo 'Requires Linux x86_64' >&2; exit 1; }
[[ $port =~ ^[1-9][0-9]{0,4}$ && $port -le 65535 ]] || { echo 'Invalid FIRESTORE_TEST_PORT' >&2; exit 1; }
for tool in curl tar sha256sum flock go; do command -v "$tool" >/dev/null || { echo "Missing dependency: $tool" >&2; exit 1; }; done
mkdir -p "$cache"
cache=$(cd "$cache" && pwd)
run=$(mktemp -d)
pid=
cleanup() {
  if [[ -n $pid ]]; then kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; fi
  rm -f "$run/emulator.log"
  rmdir "$run"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

fetch() {
  local url=$1 file=$2 digest=$3
  if [[ -f $file ]] && echo "$digest  $file" | sha256sum --check --status; then return; fi
  curl --fail --location --retry 3 --connect-timeout 15 --output "$file.part" "$url"
  echo "$digest  $file.part" | sha256sum --check --status || { echo "Checksum mismatch: $file" >&2; exit 1; }
  mv "$file.part" "$file"
}

archive="$cache/temurin-25.0.1+8-jre.tar.gz"
java="$cache/jdk-25.0.1+8-jre/bin/java"
jar="$cache/cloud-firestore-emulator-v1.19.8.jar"
(
  flock 9
  fetch 'https://github.com/adoptium/temurin25-binaries/releases/download/jdk-25.0.1%2B8/OpenJDK25U-jre_x64_linux_hotspot_25.0.1_8.tar.gz' "$archive" 126c7f59b91df97b2e930b0a456422ce29af7f2385117e1c04297eba962b0b1c
  if [[ ! -f "$cache/.jre-25.0.1+8-ready" || ! -x $java ]]; then
    tar -xzf "$archive" -C "$cache"
    "$java" -version
    touch "$cache/.jre-25.0.1+8-ready"
  fi
  fetch 'https://storage.googleapis.com/firebase-preview-drop/emulator/cloud-firestore-emulator-v1.19.8.jar' "$jar" 9d43599ed6151199e8d604dc87fac51218e49e5f3a48519b1ae560bbe5e3382d
) 9>"$cache/install.lock"

export FIRESTORE_EMULATOR_HOST="127.0.0.1:$port"
if (echo >/dev/tcp/127.0.0.1/"$port") 2>/dev/null; then
  echo "Port $port is in use. Set FIRESTORE_TEST_PORT to a free port." >&2
  exit 1
fi
"$java" -jar "$jar" --host 127.0.0.1 --port "$port" >"$run/emulator.log" 2>&1 &
pid=$!
for ((attempt=0; attempt<60; attempt++)); do
  if ! kill -0 "$pid" 2>/dev/null; then cat "$run/emulator.log" >&2; exit 1; fi
  if curl --noproxy '*' --max-time 1 -fsS "http://$FIRESTORE_EMULATOR_HOST/" >/dev/null 2>&1; then
    cd "$backend"
    if (( $# == 0 )); then set -- ./internal/business/team/adapter/firestore/group ./internal/business/team/adapter/firestore; fi
    go test -count=1 -v "$@"
    exit 0
  fi
  sleep 1
done
cat "$run/emulator.log" >&2
echo 'Firestore emulator readiness timed out' >&2
exit 1
