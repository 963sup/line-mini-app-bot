# mise ownership

Use `bash scripts/tool-exec.sh mise run <task>` from the repository root.
Tool versions, non-secret paths and tasks are defined in `mise.toml`.
The wrapper loads bootstrap-compatible environment before invoking mise.
Do not change platform HOME/CODEX_HOME or copy authentication files.

- `bootstrap`: core installation and Go doctor.
- `install-mcp`: optional local MCP tools and generated Graphify skills.
- `graphify-update` / `repomix`: refresh analysis artifacts.
- `check-config` / `doctor`: configuration and Go environment verification.
- `doctor-mcp`: configured service transports; needs network and GitHub token. Fly OAuth requires host-side verification.
- `doctor-openai`: isolated documentation transport diagnosis.
- `doctor-fly`: direct Fly transport probe; does not load host OAuth credentials.
- `check` / `build`: application checks; require an existing Go module.

Use `mise tasks` to discover current tasks. Bootstrap needs system Python 3.11+,
curl, archive/checksum utilities and network access before mise is installed.
Workspace tool paths target Codex Cloud; arbitrary machines need deliberate path
configuration. Existing cache success is not a clean rebuild test.
