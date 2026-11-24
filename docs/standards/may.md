# Optional Practices (MAY)

## Scope

These practices are permitted but not required. Their use depends on task complexity, evidence, available capabilities, applicable constraints, and expected value.

## Problem Solving

- MAY evaluate multiple implementation alternatives.
- MAY create a proof of concept to validate uncertain assumptions.
- MAY use incremental implementation to reduce integration risk.
- MAY defer nonessential improvements that do not affect acceptance criteria.
- MAY propose alternative solutions when existing assumptions are unsupported.

## Architecture

- MAY introduce additional domain layers when required by demonstrated business semantics.
- MAY extract shared capabilities when stable ownership and genuine reuse are established.
- MAY introduce new bounded contexts when independently owned business capabilities emerge.
- MAY introduce new architectural patterns when existing patterns cannot adequately satisfy validated requirements.
- MAY use feature flags to support reversible rollout or migration.

## Implementation

- MAY introduce helper functions for meaningful reuse or readability.
- MAY introduce interfaces for established capability boundaries.
- MAY introduce adapters for additional external providers.
- MAY introduce configuration options for legitimate environment or operational differences.
- MAY introduce dependencies when existing capabilities cannot adequately satisfy requirements.

## Data and Integration

- MAY introduce caching when measurable performance requirements justify it.
- MAY introduce denormalized projections for validated query requirements.
- MAY introduce asynchronous processing for operations with distinct latency or delivery requirements.
- MAY use Cloud Tasks, Pub/Sub, or Cloud Scheduler according to the operation's execution semantics.
- MAY introduce outbox or inbox patterns when reliable event delivery requires them.
- MAY use batch processing where it preserves required consistency and recovery behavior.

## API and Frontend

- MAY introduce additional API projections for demonstrated consumer requirements.
- MAY introduce feature-specific view models when generated transport types are insufficient for presentation.
- MAY use optimistic UI when operation semantics and recovery behavior support it.
- MAY introduce reusable UI components when multiple consumers share stable behavior.
- MAY use additional frontend libraries when existing capabilities are insufficient.

## Tool Usage

- MAY use Graphify, Repomix, direct source search, or language-aware tools according to the exploration question.
- MAY use additional MCP servers when they provide distinct relevant capabilities.
- MAY use local scripts for repetitive engineering operations.
- MAY use alternative verification tools when required capabilities are unavailable and acceptance requirements remain satisfied.
- MAY reuse current and applicable evidence instead of repeating equivalent investigations.

## Agent Collaboration

- MAY execute simple tasks directly.
- MAY delegate independent research, implementation, or verification tasks when supported.
- MAY use multiple agents for complex cross-module investigations.
- MAY request independent review for material architectural or security decisions.
- MAY adjust task decomposition according to discovered dependencies and risks.

## Documentation

- MAY use Mermaid diagrams when they clarify architecture, relationships, flows, or states.
- MAY create ADRs for material decisions requiring durable rationale.
- MAY use YAML front matter for documents requiring independent lifecycle tracking.
- MAY consolidate closely related documents when their authority and semantics remain intact.
- MAY create reference examples when they improve reproducibility.

## Verification

- MAY introduce benchmarks for measurable performance requirements.
- MAY introduce property-based tests for complex invariants.
- MAY introduce fuzz testing for relevant input boundaries.
- MAY use emulator-based integration tests for local service verification.
- MAY introduce end-to-end tests for critical cross-system flows.
- MAY expand verification beyond the minimum required checks when risk justifies the additional cost.

## Optimization

- MAY refactor implementations when measurable maintainability or reliability benefits exist.
- MAY remove obsolete abstractions, dependencies, and compatibility mechanisms within authorized scope.
- MAY improve nonfunctional behavior when existing acceptance criteria remain satisfied.
- MAY propose future improvements without implementing them outside the authorized task scope.

## Permission Boundary

Optional practices do not grant additional authorization and do not override mandatory requirements or prohibitions.
