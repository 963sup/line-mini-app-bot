<a id="工具與分析"></a>

# Tools and analysis

Toolchain, version management, analysis and MCP entry points.

Tool usage lives here; mise owns versions and executable tasks.

Read tool pages for concrete needs. Installed, configured, connected and successfully operated are different evidence.

<a id="工具文件目錄標準"></a>

## Tool documentation directory standards

Categorize by primary purpose. General operations belong in `shared/`, referenced by agent/MCP pages. Domain tools stay under frontend, backend, API or testing rather than moving to `agents/` for MCP support. Agent Skills/document-query MCPs belong in `agents/`.

```text
docs/development/tooling/
├── README.md
├── shared/                 # Tools shared across domains
├── frontend/               # Frontend tools
│   ├── README.md
│   ├── shadcn.md
│   └── nextjs-devtools.md
├── backend/                # Backend tools
├── api/                    # Contract, generation and validation tools
├── testing/                # Testing and quality tools
└── agents/                 # AI agent and MCP tools
```

<a id="日常工具入口"></a>

## Daily tool entry points

| Need                                                                                        | Documentation                                                                     |
| ------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| Set up tools, inspect GitHub CI, authenticate gcloud, render diagrams or generate artifacts | [Shared development tools](shared/README.md)                                      |
| Go tools/development                                                                        | [Go development](backend/go-development.md)                                       |
| Go language tools                                                                           | [gopls](backend/gopls.md)                                                         |
| Change/verify HTTP contracts                                                                | [API tools](api/change-api.md)                                                    |
| LINE Rich Menu layout                                                                       | [Grids/buttons](frontend/rich-menu-layout.md)                                     |
| Firestore tools                                                                             | [Backend tools/verification](backend/firestore.md)                                |
| Select quality checks                                                                       | [Acceptance matrix/commands](testing/commands.md)                                 |
| Open projects/submit Cloud tasks                                                            | [Codex](agents/codex.md)                                                          |
| Local container build/smoke                                                                 | [Container acceptance](../../operations/container-build.md#本機建置與-smoke-驗收) |
| Select/diagnose MCP                                                                         | [MCP index](agents/README.md)                                                     |
| Project skills                                                                              | [Skills](agents/skills.md)                                                        |
| UI components/integration                                                                   | [Frontend tools](frontend/README.md)                                              |

<a id="上游參考資料"></a>

## Upstream references

`frontend/shadcn-reference/snapshot/` is an upstream snapshot: search its index as needed rather than loading it all or treating it as project policy. [Query/update instructions](frontend/shadcn-reference/README.md)

Keep analysis artifacts in task temporary directories. Check source versions, coverage and validity; do not accumulate dated reports under docs.

[Local guidance](AGENTS.md) · [Parent guidance](../../AGENTS.md) · [Documentation entry](../../README.md)

[Parent directory guide](../README.md)
