<a id="開發手冊"></a>

# Development handbook

Read [root guidance](../../AGENTS.md), then choose the task entry. Read nested `AGENTS.md` along target paths before editing; choose checks by change type before delivery.

<a id="按任務開始"></a>

## Start by task

| Current task                                                     | Read first                                       | Next                                                                                        |
| ---------------------------------------------------------------- | ------------------------------------------------ | ------------------------------------------------------------------------------------------- |
| Prepare local/new Codex Cloud environments                       | [Environment](environment/README.md)             | Start services or run [fresh-environment acceptance](environment/codex-cloud.md#新環境驗收) |
| Start, debug, hand off or publish work                           | [Development workflow](workflow/README.md)       | Select the stage and its completion evidence                                                |
| Change APIs, domains or MINI App pages, or diagnose a deployment | [Development recipes](recipes/README.md)         | Follow the procedure for the affected surface                                               |
| Select checks/collect evidence                                   | [Development acceptance](verification/README.md) | Record results with [evidence limits](verification/evidence.md)                             |
| CLI, MCP, skills/source analysis                                 | [Tool index](tooling/README.md)                  | Select tool guides                                                                          |

<a id="歸屬與維護"></a>

## Ownership and maintenance

This handbook explains working practices. [Domains](../domains/README.md) define behavior; [architecture](../architecture/README.md) defines boundaries/contracts; the [script index](../../scripts/README.md) locates executable entries. [mise.toml](../../mise.toml) defines versions/tasks.

[Local guidance](AGENTS.md) · [Parent guidance](../AGENTS.md) · [Documentation index](../README.md)
