<a id="backendinternalbootstrap-目錄導覽"></a>

# backend/internal/bootstrap directory guide

Composes backend services, dependencies and feature routes, connecting HTTP, LINE and data adapters to business logic.

[server.go](server.go) handles startup configuration, routes, resource release and shutdown. See [line.go](line.go) for webhook ordering. Domain LINE adapters handle organization creation conversations and construction list replies.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [attendance.go](attendance.go), [creation_line.go](../business/organization/adapter/line/creation_line.go), [group.go](group.go), [miniapp.go](miniapp.go), [miniapp_test.go](miniapp_test.go), [journal.go](journal.go), [line.go](line.go), [my_engineering.go](../business/repository/adapter/line/my_engineering.go), [my_engineering_test.go](../business/repository/adapter/line/my_engineering_test.go); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)

[line_navigation.go](line_navigation.go) composes deployment-job menu synchronization with Firestore leases and bounded contexts. HTTP startup does not publish menus. Lease integration tests require an isolated loopback Firestore emulator.
