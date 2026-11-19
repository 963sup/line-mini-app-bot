# Repository scripts

Run commands from the repository root. Scripts resolve repository files from
their own location; `tool-exec.sh` also switches to the repository root.

| Directory | Responsibility |
| --- | --- |
| `tooling/` | Bootstrap, environment, configuration checks, editor wrappers, application tasks, profiling, frontend assets |
| `analysis/` | Graphify documentation graph and Graphify/Repomix scope audit |
| `api/` | OpenAPI generation, route validation and generated-code checks |
| `release/` | Release smoke checks and Cloud Run metadata inspection |

```bash
bash scripts/tooling/bootstrap.sh core
bash scripts/tooling/tool-exec.sh mise run doctor
python3 scripts/tooling/check-config.py
bash scripts/tooling/tool-exec.sh python3 -m unittest discover -s scripts/tooling -p 'test_*.py'
bash scripts/tooling/tool-exec.sh pnpm api:check
bash scripts/tooling/tool-exec.sh mise run analysis-audit
```

Tool versions and task definitions remain in `mise.toml`. Frontend-only script
paths are wired through `frontend/miniapp/package.json`; repository API commands
through root `package.json`. Release scripts do not grant deployment permission.

`tooling/shadcn-reference.py` downloads and verifies the official Components／AI
Elements documentation snapshot; it never installs application components.
See [local reference usage and refresh](../docs/development/tooling/shadcn-reference/README.md).

External bootstrap and MCP launchers use the paths under `scripts/tooling/`. Rich Menu images are stored under `assets/line/richmenu/`.

## 相關文件

- [MCP 診斷期限與清理](../docs/development/tooling/mcp/README.md)
- [文件索引：Scripts](../docs/README.md#按模組閱讀)
- [mise 任務與工具](../docs/development/tooling/mise.md)
- [分析範圍與稽核](../docs/development/tooling/analysis-scope.md)

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md)

API capability checks use a limited, comment-aware Go token reader to validate
embedded interface methods, generated signatures, adapter ownership and route
composition. Unsupported composition forms fail closed; full Go type checking
remains in the generated interface assertion and Go checks. `pnpm api:check`
runs dispatch regression tests before validating contracts and generated output.

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。
