<a id="backendinternaltransportline-工作指引"></a>

# backend/internal/transport/line agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [group.go](group.go), [reply.go](reply.go), [directory guide](./), [parent guidance](../AGENTS.md).

<a id="工作方式"></a>

## Working approach

[Current implementation diagram: basic entry points](../../../../docs/contracts/http.md#基礎入口)

[platform/line](../../platform/line/) implements group membership queries, injected through bootstrap into domain Membership ports.

## Applicable engineering conventions

- **Events and identity:** Process only verified LINE events through existing schemas and stable event identifiers; observed users do not automatically gain resource access.
- **Reliability and idempotency:** Respect reply-token lifetime and duplicate-delivery rules, with bounded outgoing calls and explicit failure handling.
- **Logging:** Correlate event handling without recording tokens or unnecessary message content.
