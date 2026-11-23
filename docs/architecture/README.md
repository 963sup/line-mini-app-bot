<a id="共用系統基礎"></a>

# Shared system foundations

Use the questions below for cross-module contracts/code. [Domain specifications](../domains/README.md) own individual features. Start with the [DDD question index](../references/ddd/README.md#按問題查閱) for models, transactions or ports.

<a id="現行架構與決策"></a>

## Current architecture and decisions

<a id="現行結構與規則"></a>

### Current structure and rules

| Question | Source |
| --- | --- |
| Which entry handles requests and which modules are affected? | [Service entry](runtime/overview.md) · [Requirements-to-code navigation](runtime/code-layout.md#從需求定位修改範圍) |
| Where are cores, adapters and composition? | [Backend boundaries](runtime/backend-boundaries.md) · [Dependencies/contract ownership](runtime/dependencies.md) |
| Who authorizes operations and which rules must hold together? | [Operation model](../contracts/business-operations.md) · [Engineering principles](../standards/engineering.md) |

<a id="任務準備與按需參考"></a>

### Task preparation and references

Select scope through the [task index](../README.md#按任務開始) and evidence through the [acceptance matrix](../development/tooling/testing/commands.md#變更驗收矩陣). Early [platform stages](../proposals/platform/delivery.md) record historical slices/open decisions; use owning contracts for current behavior.

<a id="architecture-decision-records"></a>

<a id="已接受的架構決策"></a>

### Accepted architectural decisions

Use the [ADR index](../decisions/README.md) for the adopted baseline, deployment and monorepo decisions. Record each decision’s background, options, status, impact and verification in its ADR and retain superseded decisions for history.

<a id="相鄰契約"></a>

## Adjacent contracts

<a id="http-契約"></a>

### HTTP contracts

See [public API/identity contracts](../contracts/http.md) and [cross-domain contracts](../contracts/README.md). Find change commands in the [API workflow](../../api/README.md).

<a id="身分與登入"></a>

### Identity and login

[Identity/session contract](../contracts/identity.md)

<a id="共用前端"></a>

### Shared frontend

[React page lifecycle](runtime/react-lifecycle.md) · [Runtime architecture index](runtime/README.md)

<a id="資料與儲存"></a>

### Data and storage

[Storage responsibilities](data.md) defines domain/adapter ownership. See specific [Firestore](../integrations/firestore/README.md) and [Cloud Storage](../integrations/storage/README.md) contracts.

<a id="外部整合"></a>

### External integrations

<a id="文件索引"></a>

[Integration index](../integrations/README.md) · [LINE](../integrations/line/README.md) · [Browser geolocation](../integrations/geolocation/README.md)

<a id="google-服務選用"></a>

### Selecting Google services

[Google Cloud/Maps specifications](../integrations/google-cloud/README.md) cover background jobs, functions, secrets/IAM, observability, builds, AI, reporting, Maps and reCAPTCHA by service. [Cloud Storage](../integrations/storage/cloud-storage.md) owns file storage; Firestore/Cloud Run retain existing sources. Pages state current usage/adoption criteria, not activation of candidate services.

<a id="部署與營運"></a>

### Deployment and operations

Start with [Cloud Run builds](../operations/cloud-run.md#建置入口), [release/recovery](../operations/release-acceptance.md#切換觀察回復) or [diagnostics](../operations/troubleshooting.md). Check repository declarations, actual cloud settings, builds and service acceptance separately.

<a id="可靠性設計參考"></a>

### Reliability design references

The [Webhook/Inbox/Outbox proposal](../proposals/webhook-delivery/design.md) supports retry/durable-job discussions. Determine implementation through [current HTTP contracts](../contracts/http.md) and corresponding features.

<a id="歸屬與維護"></a>

## Ownership and maintenance

Update affected entries per [local guidance](AGENTS.md) and [index maintenance](../AGENTS.md#主題子目錄) when changing architecture, integrations or operations. Find common diagrams in the [project entry point](../README.md#日常開發看哪張圖).
