<a id="backendinternaltransportline-目錄導覽"></a>

# backend/internal/transport/line directory guide

<a id="用途與範圍"></a>

## Purpose and scope

Provides LINE webhook signature validation, event transport and reply protocols.

<a id="架構與責任邊界"></a>

## Architecture and responsibility boundaries

Directories expose backend entry points or shared capabilities by purpose. Bootstrap wires services; domains own resource authorization and use cases.

[navigation](navigation) handles menu commands, postbacks and replies. Domain LINE adapters handle business commands.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [group.go](group.go), [reply.go](reply.go), [group_test.go](group_test.go), [webhook.go](webhook.go); before editing, read [local agent guidance](AGENTS.md).

<a id="閱讀入口"></a>

## Reading entry points

[Current implementation diagram: basic entry points](../../../../docs/contracts/http.md#基礎入口)

[Parent directory guide](../README.md)

[platform/line](../../platform/line/README.md) implements group membership queries, injected through bootstrap into domain Membership ports.
