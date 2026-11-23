<a id="mini-app-前端"></a>

# MINI App frontend

Browser frontend; MINI App lives in `miniapp/` and calls backend through the shared API client.

Maintain shared capabilities by layer; business owns feature UI and pages owns entry points.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [miniapp](miniapp), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Project contract](../docs/domains/project/project.md) · [Local Components/AI Elements references](../docs/development/tooling/frontend/shadcn-reference/README.md)

Projects and Repositories are sidebar peers. Project pages own group/Project selection. See [Project contracts](../docs/domains/project/project.md) for source eligibility and legacy redirects.

<a id="目錄導覽與維護"></a>

## Directory guide and maintenance

miniapp/ is the Next.js/React workspace: pages routes, business/attendance UI, business/journal journals, and identity/, platform/line/, shared/, shell/ for identity, platform, shared capabilities and framing.

Read miniapp/AGENTS.md and installed Next.js guidance first. Use @repo/api-client for HTTP; browsers obtain location and backends validate it. Do not edit build outputs.

Shared primitives live in `miniapp/src/shared/ui/`, styling in `shared/utils.ts`, business screens in `business/`. `components.json`/`tsconfig.json` define paths/aliases. Pages Router `_app.tsx` compiles `styles/globals.css` through PostCSS/Tailwind including Preflight. Use layout utilities, shadcn variants/theme variables; do not duplicate/load legacy CSS.

Follow [component protection](miniapp/src/shared/ui/AGENTS.md) and [shadcn tooling](../docs/development/tooling/frontend/shadcn.md) for UI work.

`shell/AppFrame.tsx` integrates official [sidebar-07](miniapp/src/shell/sidebar-07/README.md) for sidebar/header/mobile controls; `Shell.tsx` owns business-page identity/lifecycle. Location pages reuse AppFrame; pages do not maintain separate navigation.

Use the [matrix](../docs/development/tooling/testing/commands.md#變更驗收矩陣) for checks/builds and [API tools](../docs/development/tooling/api/change-api.md) for contract changes.

[Local guidance](AGENTS.md) · [Root README](../README.md) · [Root guidance](../AGENTS.md)

<a id="相關文件"></a>

## Related documentation

- [Production frontend entry/lifecycle](../docs/architecture/runtime/react-lifecycle.md)
- [Documentation index: Frontend](../docs/README.md#按模組閱讀)
- [Go browser sessions and sliding renewal](../docs/contracts/identity.md)
- [shadcn/MCP: components, CSS/tool verification](../docs/development/tooling/frontend/shadcn.md)
- [Browser geolocation integration](../docs/integrations/geolocation/browser-api.md)
- [Attendance queries and corrections](../docs/domains/attendance/queries.md)
- [API contract workflow](../api/README.md)

- [Group leave registration](../docs/domains/attendance/leave.md)

- [Current modules and responsibilities](../docs/architecture/runtime/code-layout.md)

<a id="business-畫面範圍"></a>

## Business screen scope

| Feature         | Existing frontend                                                     | Other entries                                                                               |
| --------------- | --------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| Journals        | `/miniapp/dashboard`/`/miniapp/feed` share JournalWorkspace           | [Design](../docs/domains/journal/journal.md)                                                |
| Attendance      | Punches, records, corrections, review/leave; records present worktime | [Attendance](../docs/domains/attendance/README.md)                                          |
| Repository      | `/miniapp/repositories` provides spaces, objects/collaboration        | [Capabilities](../docs/domains/README.md#儲存庫)                                            |
| Project         | `/miniapp/projects` provides independent cross-repository projects    | [Contract](../docs/domains/project/project.md)                                              |
| Summary/weather | LINE entry points; internal summary workers                           | [Summary](../api/contract/summary/README.md) · [Weather](../api/contract/weather/README.md) |

<a id="shadcn-頁內組合"></a>

## shadcn page composition

Correction confirmation uses Checkbox. PagePanel keeps Card radius/padding. Checkbox, Toggle and ToggleGroup use pinned registry output with project alias changes only, reusing radix-ui. sidebar-07 retains official asChild native button composition. See [attendance queries/corrections](../docs/domains/attendance/queries.md#審核畫面狀態同步) for review-list synchronization.

Attendance details use Table; corrections/reviews, leave and space lists use Item. Successful empty lists use Empty; loading/query failures remain separate. See [shadcn guidance](../docs/development/tooling/frontend/shadcn.md#頁面元件應用) for sources/purposes.

[Current implementation diagram: page and identity lifecycle](../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)
