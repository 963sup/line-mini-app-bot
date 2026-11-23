<a id="go-後端"></a>

# Go backend

Stores Go backend commands, packages and module configuration.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [package.json](package.json), [turbo.json](turbo.json) and [local guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Independent projects](../docs/domains/project/project.md) · [Engineering journals and group publication](../docs/domains/journal/journal.md)

Engineering groups and Projects each retain cores, use cases and adapters. Projects do not use Repository core models as parent containers; they read issues through consumer-defined source ports. Verify engineering-group membership and source eligibility separately. See [independent projects](../docs/domains/project/project.md) for current boundaries and migration.

<a id="目錄導覽與維護"></a>

## Directory guide and maintenance

The sole executable entry point is [main.go](main.go). [business](internal/business) is organized by domain; [bootstrap](internal/bootstrap) composes services; [identity](internal/identity) handles identity/sessions; [httpapi](internal/transport/http/api) owns OpenAPI wiring and generated code.

Separate business cores, HTTP/LINE adapters and Firestore adapters; shared composition lives in internal/bootstrap. Follow ../api contract workflows before changing HTTP formats. Do not edit generated code or put credentials in .env.example.

Use the root scripts/tooling/tool-exec.sh. Follow the [acceptance matrix](../docs/development/tooling/testing/commands.md#變更驗收矩陣) for fmt-check, lint, check, build and relevant existing tests after Go changes. Do not claim application tests for documentation-only work.

[Local guidance](AGENTS.md) · [Root README](../README.md) · [Root guidance](../AGENTS.md)

<a id="相關文件"></a>

## Related documentation

- [Engineering journals and group publication](../docs/domains/journal/journal.md)

- [Documentation index: Backend](../docs/README.md#按模組閱讀)
- [Go browser sessions and sliding renewal](../docs/contracts/identity.md)
- [Backend feature boundaries](../docs/architecture/runtime/backend-boundaries.md)
- [Dependency rules](../docs/architecture/runtime/dependencies.md)
- [Go development conventions](../docs/standards/go.md)
- [Reliability and data](../docs/proposals/webhook-delivery/design.md)
- [API contract workflow](../api/README.md)

- [Group leave registration](../docs/domains/attendance/leave.md)

- [Group attendance boards and personal missing clock-out checks](../docs/domains/attendance/board.md)

- [Current modules and responsibilities](../docs/architecture/runtime/code-layout.md)

The summary [LINE entry point](internal/business/summary/adapter/line) connects collection and commands. The [Cloud Tasks adapter](internal/business/summary/adapter/messaging/cloudtasks/summarization) schedules jobs; the [Firestore adapter](internal/business/summary/adapter/firestore) stores messages and job state.
The [CWA adapter](internal/business/weather/adapter/cwa/forecast) fetches forecasts and the [LINE adapter](internal/business/weather/adapter/line/forecast) composes replies.
See [Summary](../api/contract/summary/README.md) and [Weather](../api/contract/weather/README.md) for entry contracts.

- [Group repositories](../docs/domains/repository/README.md)
- [Basic scaffolds and expansion limits](../docs/domains/repository/acceptance.md)
- [Space label catalog](../docs/domains/repository/labels.md): reuses the Repository client, creates name bindings and deduplication receipts, and consumes no shared numbers.
- [Space discussions and comments](../docs/domains/repository/discussion/design.md): independent core and HTTP/Firestore adapters; reuse existing clients and numbering.
- [Space issues, phase one](../docs/domains/repository/issue/design.md): separate core, HTTP and Firestore layers; production acceptance pending.
- [Shared Repository resource numbering](../docs/domains/repository/numbering.md): internal transaction interface used by Issue creation; no independent public allocation API.

- [Milestones](../docs/domains/repository/milestones.md), [Project](../docs/domains/project/project.md) and [Insights](../docs/domains/repository/insights.md)

<a id="協作能力分佈與交付狀態"></a>

## Collaboration capability coverage and delivery status

Business modules live in [internal/business](internal/business); [bootstrap](internal/bootstrap) composes services. See the [feature index](../docs/domains/README.md) for behavior.

[Current implementation diagram: formats and data compatibility](../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="單一執行入口"></a>

## Single executable entry point

Start HTTP service from the root with `bash scripts/tooling/tool-exec.sh go -C backend run .`. `sync-line-navigation` is for deployment jobs; `check-firestore -document COLLECTION/DOCUMENT` runs the existing read-only connectivity check. internal/bootstrap still owns service composition.

The development-era legacy Project migration tool, its implementation, tests and operating documents are retired. Startup does not migrate data.

Go browser tests verify exported Next.js pages. Before first testing or after page changes, run `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp build` then `bash scripts/tooling/tool-exec.sh mise run check` from the root. Do not manually maintain or commit out/ HTML artifacts.
