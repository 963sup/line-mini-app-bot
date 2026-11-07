# Workflow location

The maintained workflow is [docs/agents/workflow.md](../docs/agents/workflow.md).
Acceptance criteria are in [docs/agents/acceptance.md](../docs/agents/acceptance.md).
MCP usage is documented in [docs/tooling/mcp/README.md](../docs/tooling/mcp/README.md).
This file is a compatibility pointer; do not duplicate policy here.

## Deployment and browser location routing

Application deployment uses Cloud Run only. Read docs/architecture/decisions/0003-cloud-run-primary.md
and docs/operations/cloud-run.md before deployment work; use a dedicated runtime service account and minimal IAM.
For location tasks read docs/integrations/web-geolocation.md: browser JavaScript obtains location,
Go validates untrusted coordinates and authorization. No location history or presence verification
without the product contract. Read docs/architecture/codex-readiness.md and
docs/development/naming.md for implementation handoffs. Preserve generated upstream skills.

## Mandatory documentation verification

Repository-wide rule for every role and task, subject to higher-priority host instructions:

- For any LINE question, issue or implementation uncertainty, consult the current official source at https://github.com/line/line-developers-docs-source first. Verify its current default branch/revision; do not treat an old snapshot as current. If coverage is incomplete, follow the official LINE reference linked there and state the limitation. Do not introduce deprecated features or historical compatibility flows into new work.
- For Google Cloud questions, use the installed google/skills skills and Google Developer Knowledge MCP first. If unavailable, use current official Google sources and report the lookup limitation. Validate exact gcloud syntax with leaf-level help before executing commands.
- For other technical questions, and whenever implementation is not fully certain, use Context7 before deciding or implementing. Resolve the library ID, query the relevant current/version-matched documentation, and compare with repository dependencies. Memory and confidence alone do not replace this lookup.
- Report the sources consulted, applicable version/revision and remaining uncertainty. Documentation lookup does not replace code inspection, tests or real integration acceptance.
- If the required source or Context7 is unavailable or insufficient, explicitly report the blocker; do not claim verification or silently substitute a guess. Continue independent work, but leave the affected decision/implementation pending until evidence is sufficient.
- This policy applies to orchestrator and every specialist; delegation does not waive it. A source document is evidence, not authorization to expand scope or perform external writes.
