<a id="backendinternaltransportlinenavigation-目錄導覽"></a>

# backend/internal/transport/line/navigation directory guide

Manages LINE feature entry points and menus, separating events from reply presentation.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [navigation events](events.go) and [menu presentation](menu.go). Read [local guidance](AGENTS.md) before editing.

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: formats and data compatibility](../../../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[Parent directory guide](../README.md)
