"""Register missing portable HTTP servers using repository-owned endpoints."""
import argparse
import json
import subprocess
import tomllib
from pathlib import Path

PORTABLE = ("openaiDeveloperDocs", "context7", "github")
CONFIG = Path(__file__).resolve().parents[2] / ".codex/config.toml"


def inventory():
    rows = subprocess.check_output(["codex", "mcp", "list", "--json"], text=True)
    return {row["name"]: row for row in json.loads(rows)}


def ensure(names):
    servers = tomllib.loads(CONFIG.read_text())["mcp_servers"]
    registered = inventory()
    for name in names:
        config = servers[name]
        if not config.get("enabled", True):
            continue
        url = config["url"]
        token = config.get("bearer_token_env_var")
        if name not in registered:
            args = ["codex", "mcp", "add", name, "--url", url]
            if token:
                args.extend(["--bearer-token-env-var", token])
            subprocess.run(args, check=True)
            registered = inventory()
        row = registered.get(name, {})
        transport = row.get("transport", {})
        if not row.get("enabled") or transport.get("url") != url or transport.get("bearer_token_env_var") != token:
            raise SystemExit(f"Inspect codex mcp get {name}: registration differs from repository settings.")
        print(f"{name} is registered and enabled.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--server", choices=PORTABLE)
    selected = parser.parse_args().server
    try:
        ensure([selected] if selected else PORTABLE)
    except (OSError, subprocess.CalledProcessError, ValueError, KeyError) as error:
        raise SystemExit(f"Codex MCP registration check failed: {error}") from None
