<a id="企業-工作指引"></a>

# Enterprise agent guidance

Applies to this domain; inherit [business guidance](../AGENTS.md).

<a id="建立企業的維護與驗收"></a>

## Maintaining and verifying enterprise creation

Enterprise creation must fail for authenticated users lacking provisioning eligibility. Reading details does not grant owner-change authority.

When changing this rule, check the [decision flow](../../../../docs/domains/enterprise/README.md#decision-flow) and [implementation](adapter/firestore/store.go).

[directory guide](README.md)

## Applicable engineering conventions

- **Domain, state and authorization:** Keep provisioning eligibility, enterprise ownership and team-management authority distinct, including retries returning an already-created enterprise.
