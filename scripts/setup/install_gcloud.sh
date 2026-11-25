#!/usr/bin/env bash
set -euo pipefail

if command -v gcloud >/dev/null 2>&1; then
  exec gcloud version
fi

install_root="${1:-${HOME}/.local/share}"
sdk_dir="$install_root/google-cloud-sdk"

report_installation() {
  "$sdk_dir/bin/gcloud" version
  printf 'Add to PATH for this shell: export PATH=%q:"$PATH"\n' "$sdk_dir/bin"
}

if [[ -x "$sdk_dir/bin/gcloud" ]]; then
  report_installation
  exit 0
fi

if [[ "$(uname -s)" != Linux ]]; then
  printf 'This installer supports Linux. See https://docs.cloud.google.com/sdk/docs/install-sdk\n' >&2
  exit 1
fi

case "$(uname -m)" in
  x86_64) archive=google-cloud-cli-linux-x86_64.tar.gz ;;
  aarch64|arm64) archive=google-cloud-cli-linux-arm.tar.gz ;;
  *) printf 'Unsupported Linux CPU architecture: %s\n' "$(uname -m)" >&2; exit 1 ;;
esac

mkdir -p -- "$install_root"
install_root="$(cd -- "$install_root" && pwd -P)"
sdk_dir="$install_root/google-cloud-sdk"
stage="$(mktemp -d "$install_root/.gcloud-install.XXXXXXXX")"
trap 'rm -rf -- "$stage"' EXIT

curl --fail --silent --show-error --location --retry 3 \
  --proto '=https' --proto-redir '=https' \
  --output "$stage/$archive" \
  "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/$archive"
tar -xzf "$stage/$archive" -C "$stage"
"$stage/google-cloud-sdk/bin/gcloud" version >/dev/null
mv --no-clobber --no-target-directory -- "$stage/google-cloud-sdk" "$sdk_dir"
report_installation
