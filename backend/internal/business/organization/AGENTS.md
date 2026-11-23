<a id="組織-工作指引"></a>

# Organization agent guidance

Applies to this domain; inherit [business guidance](../AGENTS.md).

<a id="移轉組織擁有權的維護與驗收"></a>

## Maintaining and verifying organization ownership transfer

Evaluate current owner/membership in one transaction. Non-owners and non-member recipients cannot complete transfers.

When changing this rule, check the [decision flow](../../../../docs/domains/organization/design.md) and [implementation](adapter/firestore/store.go).

[directory guide](./)

## Applicable engineering conventions

- **Domain, state and authorization:** Evaluate ownership and membership together with transfers/removals; legacy document translation must not grant rights from unrelated owner fields.
