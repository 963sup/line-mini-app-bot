"""Compare project MCP configuration with the effective Codex CLI inventory."""
import json
import subprocess
import tomllib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def main():
    expected = tomllib.loads((ROOT / ".codex/config.toml").read_text())["mcp_servers"]
    rows = subprocess.check_output(["codex", "mcp", "list", "--json"], cwd=ROOT, text=True)
    actual = {row["name"]: row for row in json.loads(rows)}
    failures = []
    for name, config in expected.items():
        row = actual.get(name)
        if row is None:
            failures.append(f"{name}: missing")
            continue
        transport = row.get("transport", {})
        fields = ("url", "bearer_token_env_var") if "url" in config else ("command", "args")
        if row["enabled"] != config.get("enabled", True) or any(
            transport.get(key) != config.get(key) for key in fields
        ):
            failures.append(f"{name}: enabled state or transport differs")
    if failures:
        raise SystemExit("\n".join(failures) + f"\nCheck project selection and trust for {ROOT} in Codex.")
    print(f"Matched {len(expected)} CLI registrations. Verify native tools separately in the current chat.")


if __name__ == "__main__":
    main()
