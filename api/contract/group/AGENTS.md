<a id="apicontractgroup-工作指引"></a>

# api/contract/group agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [groups.yaml](groups.yaml), the [directory guide](./) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

Follow [parent guidance](../AGENTS.md). Before changing public requests or responses, check contracts and callers, then run API contract checks. See the [directory guide](./) for implementation entry points.

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

## Applicable engineering conventions

- **Contract semantics:** Distinguish a group identifier, observed interaction candidates and verified membership in every visible group response.
