<a id="規劃與提案"></a>

# Plans and proposals

Candidate capabilities, platform targets and historical designs. Verify current capabilities through [domain status](../domains/README.md#domain-matrix), [implementation](../architecture/runtime/code-layout.md#從需求定位修改範圍) and [accepted ADRs](../decisions/README.md).

<a id="按問題查閱"></a>

## Browse by question

| Design question | Proposal/comparison |
| --- | --- |
| Durable Inbox/Outbox/Worker/admin responsibilities? | [Platform proposal](platform/README.md) · [Packages](platform/package-layout.md) · [Reliable webhooks](webhook-delivery/README.md) |
| LINE/cross-business target architecture? | [Engineering platform](line-engineering/README.md) |
| Repository/Issue/Discussion model differences? | [Resource model](platform/domain-model.md) |
| Early platform delivery order? | [Historical stages](platform/delivery.md); choose current checks using [acceptance matrix](../development/tooling/testing/commands.md). |

<a id="相鄰的設計參考"></a>

## Adjacent design references

[Firestore contracts](firestore-data-contracts.md) describes planned technical foundations. [Reliability](../architecture/README.md#可靠性設計參考) links generic Inbox/Outbox and current receiving contracts.

<a id="維護邊界"></a>

## Maintenance boundaries

Label proposal status/current authority. Move adopted rules into domains/ADRs, preserving necessary proposal rationale/references. Follow separate public-contract, production-data and deployment workflows.

Label proposal diagrams as designs; link current contracts/implementations for comparisons.

<a id="規劃中的業務領域"></a>

## Planned business domains

[Membership design](membership.md): membership, invitations, roles/revocation across users/resources.

[Operations](operations.md) · [Workflow](workflow.md) · [Agent](agent.md) · [Knowledge](knowledge.md)

[Four-domain architecture, shared vocabulary and acceptance](domain-collaboration.md)

[applicable guidance](../AGENTS.md) · [parent directory guide](../README.md)
