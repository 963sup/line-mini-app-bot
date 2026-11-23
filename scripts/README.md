# Repository scripts

Build, check, analysis/publication scripts.

Scripts own development/maintenance flows; frontend/backend packages own application rules.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Run from root. Each shell script owns one operation; bootstrap/mise compose multi-step flows. [mise.toml](../mise.toml) defines versions/tasks.

| Location                        | Responsibility                                | Task/entry                                                                                   |
| ------------------------------- | --------------------------------------------- | -------------------------------------------------------------------------------------------- |
| [bootstrap.sh](bootstrap.sh)    | Environment setup                             | core via mise run bootstrap; direct invocation defaults cloud; accepts core, cloud, mcp, all |
| [setup/](setup/README.md)       | Tools/Go dependencies                         | deps                                                                                         |
| [dev/](dev/README.md)           | Local environment/service startup             | dev                                                                                          |
| [generate/](generate/README.md) | API code, frontend assets/document graphs     | pnpm api:generate, graphify-docs                                                             |
| [check/](check/README.md)       | Configuration checks                          | check-config                                                                                 |
| [contract/](contract/README.md) | API capabilities/routes                       | pnpm api:check                                                                               |
| [test/](test/README.md)         | Route/MCP/container acceptance                | pnpm api:check, unittest, container-build                                                    |
| [perf/](perf/README.md)         | Go profiling                                  | profile                                                                                      |
| [build/](build/README.md)       | Container images                              | container-build                                                                              |
| [release/](release/README.md)   | Release acceptance                            | release-smoke.py                                                                             |
| [ops/](ops/README.md)           | Cloud Run inspection                          | cloud-run-snapshot.py                                                                        |
| [db/](db/README.md)             | Firestore indexes                             | ensure-firestore-indexes.py                                                                  |
| [tooling/](tooling/README.md)   | Development/environment diagnostics           | doctor, tool-exec.sh                                                                         |
| [mcp/](mcp/README.md)           | Project MCP settings, connections/diagnostics | check-mcp-config, doctor-mcp                                                                 |

```bash
bash scripts/bootstrap.sh core
bash scripts/tooling/tool-exec.sh mise run doctor
python3 scripts/check/check-config.py
bash scripts/tooling/tool-exec.sh python3 -m unittest discover -s scripts/test -p 'test_*.py'
bash scripts/tooling/tool-exec.sh pnpm api:check
```

Tool versions and task definitions remain in `mise.toml`. Frontend-only script
paths are wired through `frontend/miniapp/package.json`; repository API commands
through root `package.json`. Release scripts do not grant deployment permission.

`tooling/shadcn-reference.py` downloads and verifies the official Components／AI
Elements documentation snapshot; it never installs application components.
See [local reference usage and refresh](../docs/development/tooling/frontend/shadcn-reference/README.md).

Rich Menu images are stored under `assets/line/richmenu/`.

Run `mise run vuln` for Go vulnerabilities; [mise.toml](../mise.toml) pins versions.

<a id="相關文件"></a>

## Related documentation

- [MCP deadlines/cleanup](../docs/development/tooling/agents/README.md)
- [Documentation index: Scripts](../docs/README.md#按模組閱讀)
- [mise tasks/tools](../docs/development/tooling/shared/mise.md)

[Local guidance](AGENTS.md) · [Root README](../README.md)
