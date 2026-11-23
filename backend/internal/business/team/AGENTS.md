<a id="團隊-工作指引"></a>

# Team agent guidance

Applies to this domain; inherit [business guidance](../AGENTS.md).

<a id="指定團隊管理方的維護與驗收"></a>

## Maintaining and verifying team management assignment

Reassigning a managed team to another manager must conflict. Actors with target organization management permissions must also pass team-side administrator checks.

When changing this rule, check the [decision flow](../../../../docs/domains/team/README.md#decision-flow) and [implementation](adapter/firestore/store.go).

[directory guide](README.md)

## Applicable engineering conventions

- **Domain, state and authorization:** Keep team-side management checks distinct from destination-organization authority and reject conflicting management assignments.
