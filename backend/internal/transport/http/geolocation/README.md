<a id="backendinternaltransporthttpgeolocation-目錄導覽"></a>

# backend/internal/transport/http/geolocation directory guide

Provides browser-location diagnostics, separate from attendance location/authorization rules.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [geolocation.go](geolocation.go); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: attendance integration files and responsibilities](../../../../../docs/integrations/geolocation/browser-api.md#現有出勤整合檔案與責任)

[Parent directory guide](../README.md)
