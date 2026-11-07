# Workflow location

The maintained workflow is [docs/agents/workflow.md](../docs/agents/workflow.md).
Acceptance criteria are in [docs/agents/acceptance.md](../docs/agents/acceptance.md).
MCP usage is documented in [docs/tooling/mcp/README.md](../docs/tooling/mcp/README.md).
This file is a compatibility pointer; do not duplicate policy here.

## Deployment and browser location routing

Application deployment uses Fly.io only. Read docs/architecture/decisions/0001-fly-deployment.md
and docs/operations/fly.md before deployment work; Google Firestore needs explicit ADC/IAM on Fly.
For location tasks read docs/integrations/web-geolocation.md: browser JavaScript obtains location,
Go validates untrusted coordinates and authorization. No location history or presence verification
without the product contract. Read docs/architecture/codex-readiness.md and
docs/development/naming.md for implementation handoffs. Preserve generated upstream skills.
