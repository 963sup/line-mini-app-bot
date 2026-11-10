# mise ownership

## Go development loop

Run `bash scripts/tooling/bootstrap.sh core` to install the pinned core tools. In a new
shell use `bash scripts/tooling/tool-exec.sh mise run <task>`; installation alone does
not activate the shell PATH. Use `install-debug` separately for Delve.

After Go changes run `fmt-check`, `lint`, `check` and `build`. `fmt` rewrites
formatting; `fmt-check` reports differences without writing. Run `vuln` after
dependency changes and before delivery; it requires access to the vulnerability
database. CI uses these same task definitions.

Use `deadcode` as review input, not an automatic deletion rule. Supply callgraph
arguments with `mise run callgraph -- -algo=static ./cmd/server`; supply Delve
arguments with `mise run debug -- ./cmd/server`. Debugging may require ptrace
support from the host; installing Delve does not establish that permission.
Use stringer through package-specific `go:generate` directives when needed.

`coverage` writes `.artifacts/go/coverage.out`; inspect it with
`go tool cover -func=.artifacts/go/coverage.out`. `bench` runs existing benchmarks.
`profile ./path/to/package` accepts one package and writes a test binary, CPU
and memory profiles under `.artifacts/go/`. Inspect with
`go tool pprof .artifacts/go/profile.test .artifacts/go/cpu.pprof`.
These tasks do not create test cases or benchmarks. No Firestore Emulator is used.

Versions are verified against the Go module proxy; mise Go backend/exec behavior
was checked using Context7's upstream mise documentation, and coverage/benchmark
commands against Go version-matched documentation. Pins live only in mise.toml.

Use `bash scripts/tooling/tool-exec.sh mise run <task>` from the repository root.
Tool versions, non-secret paths and tasks are defined in `mise.toml`.
The wrapper loads bootstrap-compatible environment before invoking mise.
Do not change platform HOME/CODEX_HOME or copy authentication files.

- `bootstrap`: core installation and Go doctor.
- `install-mcp`: optional local MCP tools and generated Graphify skills.
- `graphify-update` / `repomix`: refresh analysis artifacts.
- `check-config` / `doctor`: configuration and Go environment verification.
- `doctor-mcp`: configured service transports; needs network and GitHub token.
- `doctor-openai`: isolated documentation transport diagnosis.
- `install-cloud`: pinned gcloud CLI and local gcloud MCP installation.
- `doctor-gcloud` / `doctor-google-knowledge`: separate Google MCP transport checks; see [Google setup](mcp/google.md).
- `check` / `build`: application checks; require an existing Go module.

Use `mise tasks` to discover current tasks. Bootstrap needs system Python 3.11+,
curl, archive/checksum utilities and network access before mise is installed.
Workspace tool paths target Codex Cloud; arbitrary machines need deliberate path
configuration. Existing cache success is not a clean rebuild test.
