# Mandatory Requirements (MUST)

## Scope

These requirements apply to all contributors, agents, modules, and engineering workflows within the repository.

## Task Execution

- MUST identify the task objective, scope, constraints, dependencies, and acceptance criteria before implementation.
- MUST read all applicable AGENTS.md files and authoritative specifications before modifying files.
- MUST inspect the current implementation before making engineering decisions.
- MUST preserve existing user changes and unrelated behavior.
- MUST identify the root cause before implementing corrective changes.
- MUST complete all authorized and feasible requirements before declaring completion.
- MUST explicitly report unresolved blockers and incomplete requirements.

## Architecture

- MUST preserve approved bounded contexts, ownership, and dependency directions.
- MUST maintain Single Source of Truth for every business rule, contract, data model, and configuration.
- MUST preserve the approved DDD and Hexagonal Architecture boundaries.
- MUST define cross-context interactions through explicit capability boundaries.
- MUST follow the authoritative architecture specifications.
- MUST document material architectural decisions and their trade-offs.

## Business and Resource Model

- MUST identify the Actor, Resource, Action, Authorization, State, and Event where applicable.
- MUST assign each business rule to its owning bounded context.
- MUST enforce business invariants before changing resource state.
- MUST define valid state transitions, preconditions, outcomes, and side effects.
- MUST preserve resource identity and lifecycle consistency.

## API Contracts

- MUST maintain OpenAPI as the authoritative HTTP contract.
- MUST follow Contract → Generate → Backend Adapter → Frontend Consumer for contract changes.
- MUST validate inputs and outputs at applicable trust boundaries.
- MUST preserve contract compatibility or provide an explicit migration.
- MUST regenerate and verify affected artifacts after contract changes.

## Security and Authorization

- MUST authenticate actors through trusted identity sources.
- MUST enforce authorization at the backend operation boundary.
- MUST validate resource ownership, membership, and access scope.
- MUST apply least privilege and deny unauthorized operations.
- MUST protect credentials, secrets, tokens, and sensitive data.
- MUST obtain explicit authorization for external or destructive operations.

## Data Consistency

- MUST define transaction and consistency requirements for state-changing operations.
- MUST protect consistency-sensitive reads and writes against concurrent modification.
- MUST define idempotency, retry, and recovery behavior where applicable.
- MUST preserve persisted data compatibility or provide an explicit migration.
- MUST publish business events only after the corresponding state change succeeds.

## Engineering Consistency

- MUST follow the established language, naming, directory, and dependency conventions.
- MUST reuse authoritative types, contracts, components, and configurations where applicable.
- MUST keep implementation, contracts, tests, and documentation synchronized.
- MUST preserve supported runtime and deployment compatibility.

## Verification

- MUST select verification methods according to the impact and risk of each change.
- MUST execute applicable formatting, linting, type, architecture, contract, and regression checks.
- MUST verify changed behavior against the original requirements and acceptance criteria.
- MUST verify applicable authorized, denied, invalid, and failure scenarios.
- MUST inspect generated changes and affected dependencies.
- MUST record actual verification results and material limitations.

## Documentation

- MUST preserve valid requirements, constraints, semantics, and decision rationale during documentation changes.
- MUST maintain one authoritative source for each specification.
- MUST update affected references when files or responsibilities change.
- MUST distinguish implemented behavior, approved specifications, proposals, and assumptions.

## Delivery

- MUST distinguish completed, verified, unverified, and blocked work.
- MUST provide evidence for material implementation and verification claims.
- MUST preserve a recoverable state before authorized high-risk operations.
- MUST follow applicable publication, deployment, and recovery workflows.
