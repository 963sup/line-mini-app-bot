# Prohibited Actions (MUST NOT)

## Scope

These prohibitions apply to all contributors, agents, modules, and engineering workflows within the repository.

## Unauthorized Actions

- MUST NOT exceed the authorized task scope.
- MUST NOT overwrite, revert, or discard unrelated user changes.
- MUST NOT perform unauthorized external writes, deletions, commits, pushes, merges, or deployments.
- MUST NOT modify protected specifications or acceptance criteria without authorization.
- MUST NOT treat tool availability as authorization.

## Agent Misconduct

- MUST NOT fabricate implementation results, verification evidence, tool usage, or subagent activity.
- MUST NOT conceal errors, regressions, incomplete work, or verification failures.
- MUST NOT silently omit required tasks or acceptance criteria.
- MUST NOT declare completion while authorized and feasible requirements remain unfinished.
- MUST NOT substitute recommendations for requested and feasible implementation.
- MUST NOT simulate unavailable tools or independent reviewers.
- MUST NOT repeatedly execute failed operations without investigating their cause.

## Architecture Violations

- MUST NOT violate approved bounded-context boundaries or dependency directions.
- MUST NOT introduce circular dependencies.
- MUST NOT access another bounded context's internal implementation or persistence directly.
- MUST NOT introduce competing architectures for existing capabilities.
- MUST NOT relocate responsibilities without preserving their ownership and contracts.
- MUST NOT introduce infrastructure, transport, framework, or external SDK dependencies into domain models.
- MUST NOT place authoritative business rules in transport adapters or infrastructure.
- MUST NOT create parallel implementations of authoritative business rules.

## Business and Resource Violations

- MUST NOT bypass business invariants or legal state transitions.
- MUST NOT perform state-changing operations through read-only queries.
- MUST NOT allow non-owning modules to independently redefine resource behavior.
- MUST NOT reuse resource identifiers with incompatible semantics.
- MUST NOT publish events representing uncommitted or nonexistent business facts.

## Authorization and Security Violations

- MUST NOT bypass authentication, authorization, signature verification, or resource-scope validation.
- MUST NOT trust client-supplied identities or permission decisions without verification.
- MUST NOT treat frontend visibility or disabled controls as authorization.
- MUST NOT expose credentials, secrets, tokens, or sensitive data.
- MUST NOT introduce insecure fallback behavior that permits unauthorized access.
- MUST NOT weaken security controls to make implementation or verification pass.

## API and Contract Violations

- MUST NOT establish independent public HTTP contracts outside OpenAPI.
- MUST NOT manually modify generated contract artifacts.
- MUST NOT introduce undocumented public operations or incompatible response semantics.
- MUST NOT expose internal persistence structures as public API contracts.
- MUST NOT introduce breaking contract changes without an explicit migration.
- MUST NOT use transport DTOs as domain entities or persistence models.

## Data and Persistence Violations

- MUST NOT bypass required transaction or consistency boundaries.
- MUST NOT separate consistency-sensitive checks from their protected writes.
- MUST NOT execute external side effects inside retried transaction callbacks.
- MUST NOT retry non-idempotent writes without duplicate-effect protection.
- MUST NOT silently discard persistence failures.
- MUST NOT modify persisted identifiers or encodings without compatibility handling.
- MUST NOT interpret an unknown operation outcome as confirmed success or failure.

## Frontend Violations

### Components

- MUST NOT create custom UI primitives when shadcn/ui provides an equivalent component.
- MUST NOT recreate functionality already provided by installed shadcn/ui components.
- MUST NOT duplicate existing components or component variants with equivalent behavior.
- MUST NOT introduce wrapper components without distinct responsibilities.
- MUST NOT introduce alternative UI component libraries without authorization.
- MUST NOT modify shared UI primitives to satisfy feature-specific requirements.

### Styling

- MUST NOT introduce custom CSS when existing Tailwind CSS utilities satisfy the requirement.
- MUST NOT hardcode colors, spacing, typography, or border values when corresponding shared theme tokens exist.
- MUST NOT introduce parallel design systems or theme definitions.
- MUST NOT duplicate styles already provided by shadcn/ui.
- MUST NOT introduce inline styles for static styling requirements covered by existing utilities.

### State

- MUST NOT store values that can be derived directly from existing state, props, or query results.
- MUST NOT duplicate server state across independent client state stores.
- MUST NOT promote feature-local state to global state without a demonstrated cross-feature requirement.
- MUST NOT use Effects solely to synchronize derived state.
- MUST NOT introduce additional state-management libraries without authorization.
- MUST NOT maintain multiple authoritative representations of the same state.
- MUST NOT allow stale responses to overwrite current actor-scoped or resource-scoped state.
- MUST NOT retain private resource data across unauthorized scope changes.

### API and Business Logic

- MUST NOT implement authoritative business rules in the frontend.
- MUST NOT duplicate backend authorization, calculations, or state transitions.
- MUST NOT manually recreate OpenAPI-generated types or API clients.
- MUST NOT bypass the established generated API client.
- MUST NOT introduce redundant API forwarding layers without distinct boundary responsibilities.
- MUST NOT expose persistence-specific models to presentation components.
- MUST NOT maintain independent frontend business models without distinct presentation requirements.

### Abstractions

- MUST NOT introduce abstraction layers without distinct responsibilities.
- MUST NOT create hooks that merely forward existing functions.
- MUST NOT create utility functions that merely rename existing functionality.
- MUST NOT introduce interfaces or types that duplicate existing contracts without semantic differences.
- MUST NOT introduce directories or modules solely to satisfy a reference structure.
- MUST NOT retain unused components, hooks, types, imports, or dependencies.
- MUST NOT introduce speculative extension points without approved requirements.

### Runtime

- MUST NOT introduce per-request server-side execution into the Next.js application.
- MUST NOT introduce Next.js API routes as alternative backend endpoints.
- MUST NOT introduce frontend-owned persistence.
- MUST NOT introduce server-side business logic into the Next.js application.
- MUST NOT introduce runtime dependencies or behavior incompatible with the Pages Router static-export deployment model.

### Minimal Implementation

- MUST NOT introduce new code that duplicates an existing capability without a distinct and necessary responsibility.
- MUST NOT introduce new implementations when existing implementations satisfy the requirements.
- MUST NOT add functionality outside approved requirements.
- MUST NOT increase architectural complexity without demonstrated necessity.
- MUST NOT retain superseded implementations after their consumers have migrated.
- MUST NOT introduce new dependencies when existing dependencies provide equivalent capabilities.
- MUST NOT sacrifice correctness, accessibility, security, or maintainability solely to reduce code size.
- MUST NOT disable required TypeScript strictness or suppress type errors to bypass validation.

## Documentation Violations

- MUST NOT create competing authoritative specifications.
- MUST NOT silently change business semantics during documentation consolidation.
- MUST NOT discard valid constraints, requirements, exceptions, or decision rationale.
- MUST NOT present assumptions, drafts, or proposals as approved specifications.
- MUST NOT describe unimplemented functionality as implemented.
- MUST NOT overwrite historical architectural decisions to conceal subsequent changes.
- MUST NOT leave superseded specifications active as competing authority.

## Verification Violations

- MUST NOT disable, bypass, or weaken required tests, lint rules, architecture checks, or CI checks.
- MUST NOT modify expected test results merely to match incorrect implementation behavior.
- MUST NOT use mocks or stubs to claim verification of untested integration behavior.
- MUST NOT treat stale generated artifacts as authoritative evidence.
- MUST NOT claim that an unexecuted check passed.
- MUST NOT substitute superficial validation for applicable behavioral verification.

## Publication and Deployment Violations

- MUST NOT bypass branch protection, required approvals, or publication workflows.
- MUST NOT publish unverified artifacts as verified releases.
- MUST NOT deploy to an unauthorized environment.
- MUST NOT perform irreversible migrations or publication without required recovery safeguards.
- MUST NOT conceal deployment failures or unresolved production risks.
