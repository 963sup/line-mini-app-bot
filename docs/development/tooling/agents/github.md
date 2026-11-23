# github

適用能力：探索、文件、規劃與自我審查。

一般儲存庫與交付狀態操作見[共用 GitHub 指南](../shared/github.md)。

## When to use

Read repository refs, issues, PRs and CI state; perform authorized collaboration
and publication using the [shared workflow](../../workflow/publication.md#repository-publication).

處理 GitHub Actions／PR 紅燈時，依[檢查失敗處理流程](../shared/github.md#github-actionspr-檢查失敗)優先使用 GitHub MCP；缺少必要能力時明列限制再採替代工具。

## Prerequisites and configuration

Use the host's GitHub MCP connection when available. The separately configured
project server receives GITHUB_MCP_TOKEN securely at runtime through config.toml;
its token presence does not establish host tool exposure or repository access.

## Operation

Tool selection, publication order, authorization and failure handling follow the
[single publication policy](../../workflow/publication.md#repository-publication).
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

## Wiki publication

頁面維護與發布核對見[共用 Wiki 操作](../shared/github.md#wiki-publication)。本頁只維護 MCP 的接線與工具限制。

## Verification and failures

Follow [MCP diagnostics](README.md#mcp-診斷) with the service name `github`.
