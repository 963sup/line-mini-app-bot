<a id="workflow-維護指引"></a>

# Workflow maintenance guidance

Applies to this directory/descendants; inherit [business guidance](../AGENTS.md). See the [directory guide](./) for scope.

Record the paused step and waiting conditions; reuse the same write identifier on resumption.

Verification: documentation checks links and git diff --check. Implementations verify no writes before confirmation and no duplicate writes after interruption following successful writes.

When behavior changes, check the [confirmation and recovery diagram](../../../../docs/proposals/workflow.md#design-diagram) and update affected branches and acceptance criteria.

## Applicable engineering conventions

- **Domain, state and authorization:** For implemented workflows, persist the resumable step and operation identity; interruption or repeated confirmation must not duplicate completed effects.
- **State and idempotency:** Define legal pause/resume/complete transitions and reuse the original write identity when resuming implemented steps.
- **Acceptance:** Keep proposed orchestration separate from deployed code and verify interruption before confirmation and after a successful side effect.
