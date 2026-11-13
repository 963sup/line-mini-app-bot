# github

Primary roles: Orchestrator, Reviewer, Integrator.

## When to use

Read repository refs, issues, PRs and CI state; perform authorized collaboration
and publication using the [shared workflow](../../agents/workflow.md#repository-publication).

## Prerequisites and configuration

Use the host's GitHub MCP connection when available. The separately configured
project server receives GITHUB_MCP_TOKEN securely at runtime through config.toml;
its token presence does not establish host tool exposure or repository access.
Connection, timeouts and enabled tool names are owned by `.codex/config.toml`.
Tool versions are owned by `mise.toml`; never copy token values into documentation.

## Operation

Tool selection, publication order, authorization and failure handling follow the
[single publication policy](../../agents/workflow.md#repository-publication).
Tool names and capabilities can differ between the host connector and project
server; use the actual exposed schemas for the intended repository and branch.

For an existing remote commit, a ref-update tool can preserve the exact SHA.
The current host's `github_update_ref` exposes `sha`, `force` and `expected_sha`;
verify its schema again if the host changes. A file-write or create-commit tool
can create a different commit even when file content matches, so it does not by
itself preserve a previously verified local commit SHA.

GitHub REST [reference updates](https://docs.github.com/en/rest/git/refs#update-a-reference)
with `force=false` require a fast-forward. This was checked through Context7
`/websites/github_en_rest`; connector-specific `expected_sha` behavior is defined
by its exposed schema, not by the REST request body.

## Verification and failures

```bash
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/mcp-doctor.py --server github
```

This checks initialize, tools/list and configured allowlists. Verify a representative
operation separately. For failures distinguish missing binary, network, credential,
workspace prerequisites and host loading; see [troubleshooting](../../operations/troubleshooting.md).
