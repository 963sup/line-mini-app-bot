# Recommended Practices (SHOULD)

## Scope

These recommendations apply across the repository unless a justified exception provides a better outcome under the applicable constraints.

## Problem Solving

- SHOULD reason from first principles.
- SHOULD identify the underlying problem rather than its symptoms.
- SHOULD prioritize high-impact changes according to evidence, risk, and cost.
- SHOULD choose the simplest solution that satisfies all requirements.
- SHOULD prefer reversible and independently verifiable changes.
- SHOULD validate uncertain assumptions through small experiments.
- SHOULD address related issues sharing the same root cause together.

## Engineering Principles

- SHOULD apply KISS, YAGNI, DRY, and separation of concerns.
- SHOULD prefer composition over inheritance or unnecessary indirection.
- SHOULD prefer explicit dependencies and behavior over implicit mechanisms.
- SHOULD favor standard solutions over custom implementations.
- SHOULD optimize for correctness, reliability, readability, and maintainability.
- SHOULD preserve reasonable extensibility without speculative implementation.

## Architecture and Domain Design

- SHOULD organize business capabilities around bounded contexts and ownership.
- SHOULD use ubiquitous language consistently across implementation and documentation.
- SHOULD model business behavior through explicit commands, queries, policies, and events where appropriate.
- SHOULD introduce aggregates, entities, and value objects when domain semantics require them.
- SHOULD keep bounded contexts internally cohesive and externally loosely coupled.
- SHOULD use the smallest interface that satisfies the caller's requirements.

## Backend Development

- SHOULD prefer idiomatic Go and the standard library.
- SHOULD prefer concrete types until an actual abstraction boundary exists.
- SHOULD define interfaces from the consumer's perspective.
- SHOULD use explicit constructors and dependency injection.
- SHOULD keep transport adapters thin and focused on translation.
- SHOULD preserve explicit error semantics and cancellation behavior.
- SHOULD use Firestore transactions or version conditions according to consistency requirements.

## API Development

- SHOULD prefer standard HTTP semantics and resource-oriented API design.
- SHOULD use generated contract types and clients.
- SHOULD provide stable resource identifiers and machine-readable error codes.
- SHOULD favor backward-compatible contract evolution.
- SHOULD define pagination, filtering, and ordering according to actual consumer requirements.

## Frontend Development

- SHOULD maintain a thin-client architecture.
- SHOULD reuse existing shadcn components and shared UI primitives.
- SHOULD use generated API types and explicit feature-level view models where necessary.
- SHOULD derive display values from existing state instead of duplicating state.
- SHOULD keep feature-specific state within its owning feature.
- SHOULD prefer accessible, responsive, and consistent interactions.

## Integration

- SHOULD isolate provider-specific behavior behind explicit integration boundaries.
- SHOULD centralize shared LINE platform capabilities.
- SHOULD translate external events into domain-specific commands or messages.
- SHOULD prefer existing Google Cloud managed services when they satisfy operational requirements.
- SHOULD use asynchronous processing when justified by latency, reliability, or delivery requirements.

## Tool Selection

- SHOULD use Graphify for architectural relationships and dependency exploration.
- SHOULD use Repomix for scoped repository and documentation analysis.
- SHOULD use gopls for Go symbol-aware navigation and diagnostics.
- SHOULD use Context7 for version-specific dependency documentation.
- SHOULD use official OpenAI developer documentation for OpenAI and Codex behavior.
- SHOULD use GitHub MCP for repository and pull request operations.
- SHOULD use existing Skills and repository scripts before introducing new workflows.
- SHOULD corroborate tool-derived findings against current authoritative sources.

## Agent Collaboration

- SHOULD delegate independent complex tasks when supported.
- SHOULD execute independent tasks in parallel.
- SHOULD execute dependent tasks in dependency order.
- SHOULD assign explicit scope, deliverables, and acceptance criteria to subagents.
- SHOULD use independent review for high-risk or cross-module changes.
- SHOULD consolidate evidence before making final engineering decisions.

## Documentation

- SHOULD organize specifications by authority, responsibility, and subject.
- SHOULD prefer concise, declarative, and verifiable requirements.
- SHOULD use references instead of duplicating authoritative content.
- SHOULD use diagrams when they materially improve understanding.
- SHOULD consolidate related documentation without losing semantics.

## Verification and Optimization

- SHOULD automate recurring mechanical verification.
- SHOULD prioritize regression checks for high-risk business behavior.
- SHOULD establish measurable baselines before optimization.
- SHOULD compare results before and after material changes.
- SHOULD prioritize removing complexity before introducing optimization mechanisms.
- SHOULD stop further optimization when additional changes lack demonstrated value.
