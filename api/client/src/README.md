<a id="apiclientsrc-目錄導覽"></a>

# api/client/src directory guide

Provides client creation and request credential configuration. Types in `generated/` are generated from public OpenAPI contracts; change fields at the contract source.

Contracts define interfaces; consumers use the client and generation settings to regenerate contract types.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [client.ts](client.ts) and [generated](generated). Read the [local agent guidance](AGENTS.md) before editing.

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: basic entry points](../../../docs/contracts/http.md#基礎入口)

[Parent directory guide](../README.md)
