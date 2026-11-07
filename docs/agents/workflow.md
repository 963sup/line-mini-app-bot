# Go delivery workflow

Eight roles are available; they are not eight mandatory execution stages.
The orchestrator can complete small work directly. Use implementer, tester and
reviewer for substantial changes; add architect/researcher/explorer only for a
specific unresolved question. Use integrator when separate patches require it.
Start with two or three independent workers, within host limits. One writer owns
each file; test fixtures must have explicit ownership too.

## Assignment contract

Provide the goal, owned files, dependencies, behavior to preserve and acceptance
commands. Return changed paths, actual commands/results, skips and open issues.
Do not repeat an investigation already answered unless new evidence conflicts.


## Reading and handoff

Read [acceptance](acceptance.md) before choosing checks. The root AGENTS.md routes
roles to relevant documents. Resolve routine issues within the authorized scope;
report host restrictions accurately. Consult [MCP guidance](../tooling/mcp/README.md)
for tool selection and transport diagnostics. Integration does not grant permission
to push or deploy.
