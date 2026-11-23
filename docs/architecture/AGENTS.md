<a id="共用系統基礎文件指引"></a>

# Shared system foundation documentation guidance

Follow [documentation/verification guidance](../AGENTS.md); select cross-module contracts, integrations or decisions through the [index](./).

## architecture

Covers runtime, dependencies and storage responsibilities; shared identity/authorization live in [cross-boundary contracts](../contracts/).

Cross-module boundaries, dependency direction, HTTP/data contracts and architectural decisions.
Keep local commands in module READMEs, feature rules in domains and deployment procedures in operations. Accepted ADRs in decisions are authoritative; clearly label future plans.

Use the [baseline](../decisions/0002-baseline.md) and [foundation specification](../development/environment/verification.md) for stage scope: infrastructure, LINE/Firestore, then attendance. Platform expansion is deferred reference; Worker, Garr, location and new testing infrastructure are not requirements for this stage.

<a id="外部整合"></a>

## External integrations

The [integration index](../integrations/) covers LINE, geolocation, Firestore, Cloud Storage and Google Cloud. See [Firestore integration](../integrations/firestore/connection.md) for storage rules.
Separate external-service contracts from project feature rules. Follow root verification policies; store no tokens, credentials or secrets in documents.

## operations

Follow [operations guidance](../operations/AGENTS.md) for publication/acceptance.

## Applicable engineering conventions

- **Boundaries and dependencies:** Describe owned data, dependency direction and integration boundaries using the current module structure and contracts.
- **Consistency and evolution:** Make transaction, concurrency, event and migration assumptions explicit; record consequential decisions through the existing ADR process.
- **Evidence:** Label current/proposed diagrams and link their implementation or acceptance sources, rather than presenting a target architecture as deployed behavior.
