<a id="frontendminiapp-目錄導覽"></a>

# frontend/miniapp directory guide

<a id="用途與範圍"></a>

## Purpose and scope

Stores the LINE MINI App Next.js application, build settings/tests.

<a id="架構與責任邊界"></a>

## Architecture and responsibility boundaries

`pages/` composes pages; `business/` owns feature forms/screens; `shared/ui/` supplies shadcn. OpenAPI defines exchanged data. Backends authorize every read/write against current identity/persistence state.

| Feature | Backend supplies | Browser owns |
|---|---|---|
| Attendance/corrections/leave | Preflight, monthly worktime, review eligibility/actions, requests | Location, unsent time edits, rejection reasons, loaded records |
| Discussions/comments/issues/milestones | Edit capabilities, classification, counts/completion | Drafts, labels, progress/keyboard interaction |
| Objects/subobjects | Quantities/transitions/reset conditions | Unsent decimal validation |
| Projects/repositories | Sources, references/inspection modes, versions, statistics/Wiki history | Deep links, pagination, source rechecks, grouped references/Wiki formatting |
| Journals | Reference candidates, work snapshots, actual/expected cumulative progress | Group selection, work rows/group charts, cancel stale requests |
| Enterprises/organizations/teams/groups | Provisioning/management eligibility, valid organizations, teams/membership | Share organization queries during sessions and navigate after selection; attendance/journals/projects share group queries with separate pagination |
| Identity/navigation/diagnostics | LINE verification, sessions/location validation | LIFF/browser APIs, invalid-identity cleanup, sidebar/focus/cancellation |

Tests cover identity switches, retries, stale responses, pagination/edit conflicts. Check domain rules, responses and controls together when changing permissions/work state.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [components.json](components.json), [next-env.d.ts](next-env.d.ts), [next.config.ts](next.config.ts), [package.json](package.json), [postcss.config.mjs](postcss.config.mjs), [src](src); before editing, read [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

[Parent directory guide](../README.md)

<a id="下層文件入口"></a>

## Child documentation entry points

- [tests](tests/README.md)
