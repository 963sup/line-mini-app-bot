<a id="line-群組相容整合"></a>

# LINE group compatibility integration

[domain documentation](../../../docs/domains/group/README.md) · [global index](../../../docs/domains/README.md#domain-matrix)

Stores group data/member candidates from signature-verified LINE events and provides group queries.

Owns LINE group observation, legacy group queries and compatibility APIs. Team owns state/ownership; Membership owns formal eligibility. Candidate indexes aid discovery only; reads still require live LINE membership verification.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [group.go](application/query/group.go), [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Project design](../../../docs/domains/project/project.md) · [Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="閱讀入口"></a>

## Reading entry points

Group uses LINE GroupID directly as document ID and stores verified events/member candidates. See [Project design](../../../docs/domains/project/project.md) for Project ownership/authorization.

[Current implementation diagram: formats and data compatibility](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

<a id="領域追蹤入口"></a>

## Domain traceability entry points

[domain specification](../../../docs/domains/group/README.md) · [decision flow](../../../docs/domains/group/README.md#decision-flow) · [representative tests](application/query/group_test.go)

[Parent directory guide](../README.md)
