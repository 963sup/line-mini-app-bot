<a id="api-能力與路由契約"></a>

# API capability and route contracts

Entry: pnpm api:check.

- [check-capabilities.mts](check-capabilities.mts)
- [check-dispatch.mts](check-dispatch.mts)
- [check-compatibility.mts](check-compatibility.mts): run `bash scripts/tooling/tool-exec.sh pnpm api:compat --base COMMIT` with an explicit existing Git revision. It bundles and dereferences public/internal contracts from that revision and the working tree using the pinned Redocly tool.

The historical gate conservatively rejects semantic changes to existing operations, path parameters, authentication schemes, global security and servers. New operations and documentation edits are allowed. It may reject a compatible extension to an existing operation: this is a review gate, not a complete OpenAPI subtyping proof. CI compares PRs with their merge base and pushes with the previous head; a first push has no historical baseline. Missing history or unsupported bundling fails rather than silently skipping. Keep `api:check` for current schema/consumer/generated-code validation; never reset a baseline to conceal a change.

[Script overview](../README.md) · [Local guidance](AGENTS.md)
