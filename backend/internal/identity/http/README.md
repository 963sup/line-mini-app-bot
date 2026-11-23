<a id="backendinternalidentityhttp-目錄導覽"></a>

# backend/internal/identity/http directory guide

Connects authentication/sessions to HTTP, handling login and request identity extraction.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [identity.go](identity.go), [session.go](session.go); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: renewal transaction and frontend triggers](../../../../docs/contracts/identity.md#續期交易與前端觸發)

[Parent directory guide](../README.md)
