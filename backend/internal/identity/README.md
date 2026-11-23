<a id="backendinternalidentity-目錄導覽"></a>

# backend/internal/identity directory guide

Defines shared identity/owner verification, integrating LINE/Google credential validation and application sessions.

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [identity.go](identity.go), [owner.go](owner.go), [google](google), [group_owner_test.go](group_owner_test.go), [http](http), [line](line); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: renewal transaction and frontend triggers](../../../docs/contracts/identity.md#續期交易與前端觸發)

[Parent directory guide](../README.md)

<a id="下層文件入口"></a>

## Child documentation entry points

- [session](session/README.md)
