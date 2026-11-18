"""Render the supported non-secret mise environment before mise itself starts."""
import argparse
import os
from pathlib import Path
import re
import shlex
import sys

if sys.version_info < (3, 11):
    raise SystemExit("Bootstrap requires Python 3.11+ with tomllib before mise is installed.")
import tomllib

NAME = re.compile(r"[A-Za-z_][A-Za-z0-9_]*")
DEFAULT = re.compile(r'{{ env\.([A-Za-z_][A-Za-z0-9_]*) \| default\(value="([^"]*)"\) }}')


def mise_version(config):
    """Use the minimum supported mise release as the bootstrap version."""
    version = config.get("min_version")
    if not isinstance(version, str) or not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", version):
        raise ValueError("mise.toml min_version must be an exact numeric release")
    return version


def render_exports(settings, environment):
    """Validate the whole table before returning shell-quoted export statements."""
    if not isinstance(settings, dict):
        raise ValueError("mise.toml [env] must be a table")
    exports = []
    for name, value in settings.items():
        if not NAME.fullmatch(name):
            raise ValueError("Invalid environment variable name in mise.toml [env]")
        if not isinstance(value, str):
            raise ValueError(f"Unsupported bootstrap env value: {name}; expected string")
        match = DEFAULT.fullmatch(value)
        if match:
            value = environment.get(match[1]) or match[2]
        elif "{{" in value or "}}" in value:
            raise ValueError(f"Unsupported bootstrap template: {name}")
        if "\0" in value:
            raise ValueError(f"Invalid NUL byte in environment variable: {name}")
        exports.append(f"export {name}={shlex.quote(value)}")
    return "\n".join(exports)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--mise-version", action="store_true", help="Print the validated bootstrap release")
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    try:
        config = tomllib.loads((root / "mise.toml").read_text())
        exports = mise_version(config) if args.mise_version else render_exports(config["env"], os.environ)
    except (OSError, KeyError, ValueError) as error:
        # Do not print configuration contents or inherited environment values.
        if isinstance(error, tomllib.TOMLDecodeError):
            raise SystemExit("Invalid mise.toml syntax") from None
        raise SystemExit(f"Cannot load development environment: {error}") from None
    print(exports)


if __name__ == "__main__":
    main()
