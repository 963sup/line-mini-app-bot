<a id="api-能力與路由契約工作指引"></a>

# API capability and route contract guidance

Inherit [parent guidance](../AGENTS.md); after contract-check changes, run pnpm api:check and verify generated code/routes.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **API and generation:** Verify operation ownership, route dispatch, registered consumers and generated signatures against the same contract revision.
- **Compatibility:** Update consumer inventory when a call is added or removed while retaining supported public operations and stable errors.
- **Acceptance:** Reproduce invalid contract/dispatch cases and preserve fail-closed checks; generated output is verified rather than patched.
