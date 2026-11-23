<a id="frontendminiappsrc-工作指引"></a>

# frontend/miniapp/src agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [business](business), [diagnostics](diagnostics), [identity](identity), [navigation](navigation), [directory guide](README.md), [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

<a id="驗證入口"></a>

## Verification entry points

Run `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check` from root for types. For flows/interactions, replace `check` with `test` and verify affected identity, failures/retries. Verify LINE WebView flows in their target environments separately.

## Applicable engineering conventions

- **Structure:** Keep routes, identity, platform integration, business UI and shared mechanics in their existing responsibility directories.
- **Types and dependencies:** Import generated contracts through the shared API client; avoid cross-feature imports of private state or duplicate transport types.
- **Lifecycle:** Cancel obsolete reads and clear identity/resource-scoped data rather than allowing an old response to overwrite a new scope.
