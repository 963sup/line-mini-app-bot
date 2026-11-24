# Discouraged Practices (SHOULD NOT)

## Scope

These practices should be avoided unless a justified exception demonstrates sufficient benefit relative to their cost and risk.

## Complexity

- SHOULD NOT introduce speculative functionality.
- SHOULD NOT design for hypothetical requirements without supporting evidence.
- SHOULD NOT introduce abstractions before concrete shared requirements exist.
- SHOULD NOT introduce additional layers solely to satisfy a reference directory structure.
- SHOULD NOT replace straightforward implementations with generic frameworks without demonstrated benefit.
- SHOULD NOT introduce configuration options for behavior that does not require variability.
- SHOULD NOT optimize for theoretical scalability at the expense of current correctness and maintainability.

## Architecture

- SHOULD NOT split cohesive business capabilities into unnecessarily small modules.
- SHOULD NOT merge unrelated capabilities solely to reduce package counts.
- SHOULD NOT introduce shared business packages without clear ownership.
- SHOULD NOT introduce intermediate services that merely forward calls without adding boundary responsibilities.
- SHOULD NOT create cross-context abstractions that obscure business semantics.
- SHOULD NOT introduce distributed services when an existing modular deployment satisfies requirements.
- SHOULD NOT introduce hidden dependencies through global mutable state or service locators.
- SHOULD NOT move domain-specific behavior into shared packages merely to avoid bounded-context ownership.
- SHOULD NOT expose internal domain implementations as cross-context contracts when an explicit capability contract is sufficient.

## Implementation

- SHOULD NOT duplicate equivalent functions, components, or utilities without justified differences.
- SHOULD NOT introduce interfaces with only speculative consumers.
- SHOULD NOT introduce wrapper functions that add no semantic or technical value.
- SHOULD NOT create unnecessary mapping layers between equivalent representations.
- SHOULD NOT use reflection, dynamic dispatch, or metaprogramming where explicit code is sufficient.
- SHOULD NOT retain obsolete compatibility code after its supported consumers have been retired.
- SHOULD NOT introduce dependencies for capabilities adequately provided by existing dependencies or standard libraries.

## Backend and Persistence

- SHOULD NOT introduce repositories or services that merely duplicate lower-level interfaces.
- SHOULD NOT use generic repositories when domain-specific access semantics are required.
- SHOULD NOT introduce caching before identifying a measurable performance bottleneck.
- SHOULD NOT introduce denormalized data without demonstrated query or consistency requirements.
- SHOULD NOT introduce asynchronous messaging for operations adequately served by synchronous execution.
- SHOULD NOT introduce distributed transactions when simpler consistency boundaries are sufficient.

## Distributed Execution

- SHOULD NOT assume events or background jobs are delivered exactly once or in order unless the infrastructure provides an enforceable guarantee.
- SHOULD NOT regenerate an operation identity when retrying the same business operation.
- SHOULD NOT acknowledge durable work before its required persistence or processing guarantee is satisfied.
- SHOULD NOT introduce unbounded retries, polling, or recursive execution.
- SHOULD NOT rely on process-local memory or instance-local locks as the sole authority for durable distributed state or invariants.
- SHOULD NOT rely on a continuously running service instance for required durable execution.
- SHOULD NOT allow stale workers to overwrite results produced by newer operation versions.
- SHOULD NOT remove recovery records while an operation outcome remains unresolved.
- SHOULD NOT silently discard permanently failed background operations.

## API Design

- SHOULD NOT expose unnecessary fields or operations.
- SHOULD NOT introduce custom protocols when standard HTTP semantics are sufficient.
- SHOULD NOT create endpoint-specific transport conventions without a distinct requirement.
- SHOULD NOT introduce API versions solely for internal implementation changes.
- SHOULD NOT duplicate equivalent schemas solely for naming convenience.

## Frontend Development

- SHOULD NOT introduce global state for feature-local behavior.
- SHOULD NOT use Effects for values directly derivable during rendering.
- SHOULD NOT create custom components when existing components satisfy the requirement.
- SHOULD NOT introduce independent design tokens for existing semantic roles.
- SHOULD NOT introduce client-side caching or synchronization mechanisms without demonstrated need.
- SHOULD NOT introduce unnecessary client dependencies or state-management frameworks.

## Agent Execution

- SHOULD NOT perform repository-wide exploration for narrowly scoped tasks.
- SHOULD NOT repeatedly generate unchanged repository snapshots.
- SHOULD NOT reread unchanged evidence without a concrete verification need.
- SHOULD NOT delegate simple tasks when direct execution is clearer and faster.
- SHOULD NOT assign overlapping work that creates avoidable conflicts or duplicate effort.
- SHOULD NOT continue exploration after sufficient evidence supports the decision.
- SHOULD NOT redefine task scope or acceptance criteria to justify incomplete work.
- SHOULD NOT present partial completion or a temporary implementation as the final result.
- SHOULD NOT use TODOs, placeholders, or hardcoded responses to claim that functional requirements are satisfied.
- SHOULD NOT treat subagent conclusions as verified facts without cross-validation.
- SHOULD NOT treat generated summaries, code graphs, or cached analysis as substitutes for current source evidence.

## AI Agent and Knowledge

- SHOULD NOT treat AI-generated output or model confidence as verified evidence.
- SHOULD NOT publish generated content as authoritative knowledge without traceable evidence and applicable review.
- SHOULD NOT overwrite original evidence during knowledge extraction, consolidation, or distillation.
- SHOULD NOT resolve conflicting evidence by silently discarding contradictory sources.
- SHOULD NOT reuse knowledge outside its authorized resource or tenant scope.
- SHOULD NOT use revoked, expired, or superseded knowledge as active guidance.

## Integration and Resources

- SHOULD NOT introduce duplicate ownership of LINE commands or postback actions.
- SHOULD NOT introduce independent LINE Messaging API clients outside approved composition boundaries.
- SHOULD NOT treat provider acceptance as proof of message delivery or user receipt.
- SHOULD NOT create orphaned or cyclic relationships in hierarchical resource models.
- SHOULD NOT mutate published template revisions in place.
- SHOULD NOT retroactively modify instantiated objects solely because their source template changed.
- SHOULD NOT infer missing quantities, units, or contractual requirements as verified source facts.
- SHOULD NOT remove source provenance from derived business objects without an explicit retention decision.

## Documentation

- SHOULD NOT duplicate authoritative requirements in navigation or overview documents.
- SHOULD NOT create documents without distinct ownership, purpose, and audience.
- SHOULD NOT use diagrams when concise prose communicates the same relationship more clearly.
- SHOULD NOT retain superseded guidance after its references and consumers have migrated.
- SHOULD NOT mix approved requirements, proposals, examples, and historical notes without clear labels.

## Governance

- SHOULD NOT introduce local exceptions that weaken repository-wide guidance without recording their scope, rationale, impact, owner, and approval.
- SHOULD NOT modify governing rules merely to legitimize a noncompliant implementation.
- SHOULD NOT introduce contradictory requirements without explicitly resolving their authority.
- SHOULD NOT treat examples, generated artifacts, or implementation details as overriding authoritative specifications.

## Verification and Optimization

- SHOULD NOT run unrelated expensive checks when narrower verification provides sufficient evidence.
- SHOULD NOT rely only on snapshots, type checks, or mocks for behavior requiring integration evidence.
- SHOULD NOT optimize without a measurable baseline and acceptance target.
- SHOULD NOT retain an optimization whose complexity exceeds its demonstrated value.
- SHOULD NOT expand a fix into unrelated cleanup without authorization or a shared root cause.
