<a id="agent-維護指引"></a>

# agent maintenance guidance

Applies to this directory/descendants; inherit [business guidance](../AGENTS.md). See the [directory guide](./) for scope.

Declare allowed operations, readable resources and step limits for each task. Validate model-proposed calls before execution.

Verification: for documentation, check links and git diff --check. For implementations, verify unauthorized operations are rejected and tasks terminate with results when steps or evidence are exhausted.

When changing this domain's behavior, check the [investigation and termination diagram](../../../../docs/proposals/agent.md#design-diagram) and update affected branches and acceptance criteria.

## Applicable engineering conventions

- **Domain, state and authorization:** Validate model-proposed operations against the task allowlist, resource scope and step budget before executing side effects.
- **Reliability and lifecycle:** Bound model/tool calls, preserve cancellation and stop at the declared task budget; retries must not expand the operation allowlist.
- **Verification:** For implemented behavior, test denied operations, exhausted evidence/budgets and termination outcomes without inventing persistent workflow infrastructure.
