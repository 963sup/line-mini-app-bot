# workflow

[domain documentation](../../../../docs/domains/workflow/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Manages explicit step ordering, waiting conditions and progress.

[Confirmation/resumption use cases](application/confirmation.go) define workflow storage and business-command interfaces. Persist actors, candidates and submission IDs before resumption. Commands reauthorize and return reusable receipts. Persistence adapters and product entry points remain unwired.

<a id="共用接線"></a>

## Shared wiring

Use [Preparation](application/prepare.go) in named business use cases to connect existing read-only Operations, optional Knowledge/Agent investigation and result assembly. Trusted callers supply actors/targets. Program conditions choose optional steps; return immediately on failure. Read-only needs may call original queries directly or Preparation without optional steps.

For human confirmation, callers persist workflows and immutable candidates, then invoke `Stage` with the workflow version. Order: `Prepare → Stage → Confirm → Resume`. Preparation itself writes no business data. Domains wire named reads, candidate assembly and existing commands; generic flows do not depend on journal, attendance or document types.

[Cross-capability tests](application/prepare_test.go) use the existing Investigator and Operations/Knowledge contracts to cover skipping optional steps, sequential investigation for incomplete data and stopping after denied reads. Confirmation/resumption uses test doubles; no product use cases or persistence adapters are configured.

[Confirmation and recovery diagram](../../../../docs/proposals/workflow.md#design-diagram) · [Maintenance guidance](AGENTS.md)

[Parent directory guide](../README.md)
