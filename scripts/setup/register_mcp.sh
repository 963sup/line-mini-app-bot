#!/usr/bin/env bash
set -euo pipefail

setup_dir=$(dirname -- "${BASH_SOURCE[0]}")
bash "$setup_dir/register_gopls.sh"
bash "$setup_dir/install_shadcn.sh" --register-only
bash "$setup_dir/install_next_devtools.sh" --register-only
bash "$setup_dir/register_openai_developer_docs.sh"
bash "$setup_dir/register_context7.sh"
bash "$setup_dir/register_github.sh"

codex mcp list
