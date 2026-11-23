<a id="knowledge-維護指引"></a>

# Knowledge maintenance guidance

Applies to this directory and descendants; inherit [business guidance](../AGENTS.md). Check the [knowledge version lifecycle](../../../../docs/proposals/knowledge.md#design-diagram) when changing publication, queries or withdrawal.

Filter queries by trusted identity, organization and scope. Verify drafts/withdrawn versions are excluded from new tasks and entries are inaccessible to other organizations. For documentation, check relative links and git diff --check.

[directory guide](./)

## Applicable engineering conventions

- **Domain, state and authorization:** For implemented publication/read paths, authorize scope and version visibility explicitly; drafts or withdrawn knowledge must not become silently available to new tasks.
- **Version and cache lifecycle:** For implemented readers, scope cached knowledge by authorized version and invalidate it when publication or withdrawal changes visibility.
- **Acceptance:** Keep proposal-only policies separate from deployed behavior and test cross-organization denial whenever a reader is implemented.
