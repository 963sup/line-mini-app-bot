<a id="專案文件入口"></a>

# Project documentation entry point

Select one authoritative entry by task, then follow its contracts, code and representative tests.

<a id="按任務開始"></a>

## Start by task

| Task | Read first | Then as needed |
| --- | --- | --- |
| Locate changes | [Requirements-to-code navigation](architecture/runtime/code-layout.md#從需求定位修改範圍) | Feature specifications and directory AGENTS.md |
| Change behavior/authorization | [Feature index](domains/README.md) | [Operation model](contracts/business-operations.md), [contract ownership](architecture/runtime/dependencies.md#契約與不變條件的維護位置) |
| Change public APIs | [API contracts/generation](../api/README.md) | Relevant feature and [HTTP contracts](contracts/http.md) |
| Define models, aggregates or transactions | [DDD question index](references/ddd/README.md#按問題查閱) | [Current backend boundaries](architecture/runtime/backend-boundaries.md) and domain tests |
| Edit owned documents/guidance | [Documentation ownership](#文件歸屬與維護) | [Maintenance/verification](AGENTS.md#驗證) |
| Prepare development | [Setup guide](development/tooling/shared/toolchain-setup.md) | [Tools/MCP](development/tooling/README.md) |
| Determine completion | [Acceptance matrix](development/tooling/testing/commands.md#變更驗收矩陣) | Feature and external-environment acceptance |
| Deploy, publish or troubleshoot | [Operations](operations/README.md) | Configuration and real-environment evidence |
| Missing Codex tools | [MCP loading checks](development/tooling/agents/README.md#project-loading-check) | [Project Codex configuration](../.codex/config.toml) |
| Select Google Cloud, Maps or reCAPTCHA | [Service specifications](integrations/google-cloud/README.md) | [Storage](architecture/data.md) and domain contracts |
| Assess future capabilities | [Plans/proposals](proposals/README.md) | Accepted ADRs and current specifications |

<a id="日常開發看哪張圖"></a>

## Diagrams for daily development

Start with runtime entry points when taking over; otherwise choose relevant sections by question. Implementation, deployment and device acceptance status depend on each page's evidence.

| Development question | Current implementation diagram |
| --- | --- |
| Where do browser, LINE and Cloud Tasks requests converge? | [Runtime entry points](architecture/runtime/overview.md#執行入口) |
| When is identity rechecked after returning from the background? | [Page/identity lifecycle](architecture/runtime/react-lifecycle.md#頁面與身分生命週期) |
| How do LINE tokens become sessions, and where are resources authorized? | [MINI App trust flow](contracts/http.md#mini-app-信任流程) |
| When do webhooks succeed or request redelivery? | [Webhook processing](contracts/http.md#基礎入口) |
| Where are API/Firestore field formats converted? | [Backend data boundaries](architecture/runtime/backend-boundaries.md#格式與資料相容) |
| How do projects reference sources across Repositories? | [Projects/source references](domains/project/project.md#最小來源引用骨架) |

Other diagrams by purpose:

| Purpose | Entry points |
| --- | --- |
| Current business flows/states | [Domain index](domains/README.md#domain-matrix) · [Location punch sequence](domains/attendance/README.md#定位前預檢狀態與時序) · [Corrections/manual review](domains/attendance/worktime.md#補卡與人工審核) |
| Models, aggregates, transactions/use cases | [DDD diagrams](references/ddd/design-diagrams.md) · [Aggregate boundaries](references/ddd/aggregate.md#邊界與存取入口) · [Commit/publication](references/ddd/transaction.md#提交與發布) |
| Development, acceptance/troubleshooting | [Development workflow](development/README.md) · [Acceptance matrix](development/tooling/testing/commands.md#變更驗收矩陣) · [Deployment/recovery](operations/release-acceptance.md#切換觀察回復) |
| Unadopted designs | [Proposals](proposals/README.md) · [Cross-domain collaboration](proposals/domain-collaboration.md) · [Webhook/Inbox/Outbox](proposals/webhook-delivery/design.md#webhook-接收處理與發送) |

<a id="目錄責任"></a>

## Directory responsibilities

| Category | Authority and reading approach |
|---|---|
| [Architecture](architecture/README.md) | System boundaries, trust, storage and runtime architecture. |
| [Contracts](contracts/README.md) | Cross-boundary identity, authorization, data and operations. |
| [References](references/README.md) | DDD concepts, design methods and external sources. |
| [Integrations](integrations/README.md) | LINE, geolocation, Firestore, Cloud Storage and Google service indexes. |
| [Domains](domains/README.md) | Rules, core flows and acceptance across 18 domains. |
| [Decisions](decisions/README.md) | Accepted ADR backgrounds/reasons; cite by number. |
| [Development](development/README.md) | Change workflows, tools, debugging and acceptance. |
| [Operations](operations/README.md) | Production deployment, release acceptance, troubleshooting/recovery. |
| [Proposals](proposals/README.md) | Designs and adoption criteria under evaluation. |
| [Standards](standards/README.md) | Naming, code, documentation and delivery rules. |

<a id="按模組閱讀"></a>

## Read by module

[Backend](../backend/README.md) · [API](../api/README.md) · [Frontend](../frontend/README.md) · [Scripts](../scripts/README.md) · [Assets](../assets/README.md) · [Production operations](operations/README.md)

<a id="文件歸屬與維護"></a>

## Documentation ownership and maintenance

Domains own feature rules, architecture owns shared mechanisms, and [ADRs](decisions/README.md) own accepted major decisions. Plans do not override current contracts. Keep historical changes in Git rather than old reports under docs; link complete rules without duplicating policies/versions.

Update indexes, references, relative paths and anchors for new/moved pages; verify per [documentation standards](standards/documentation.md). [mise.toml](../mise.toml) and workspace configuration define tool versions/commands.

[Parent directory guide](../README.md)
