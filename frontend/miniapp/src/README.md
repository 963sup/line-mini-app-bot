<a id="frontendminiappsrc-目錄導覽"></a>

# frontend/miniapp/src directory guide

Use **feature-based architecture**: `pages/` routes/composes pages, `business/` owns business screens/interactions, `shell/` shared layout/navigation and `shared/` cross-feature utilities/components. **Split features as needed.** Simple screens may remain one component; separate forms/state/requests when independently responsible.

Maintain shared capabilities by layer; business owns feature UI and pages owns entry points.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [business](business), [diagnostics](diagnostics), [identity](identity), [navigation](navigation), [pages](pages), [platform](platform); before editing, read [local agent guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Current implementation diagram: page and identity lifecycle](../../../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期)

<a id="驗證入口"></a>

## Verification entry points

Run `bash scripts/tooling/tool-exec.sh pnpm --filter @repo/miniapp check` from root for types. For flows/interactions, replace `check` with `test` and verify affected identity, failures/retries. Verify LINE WebView flows in their target environments separately.

[Parent directory guide](../README.md)

<a id="下層文件入口"></a>

## Child documentation entry points

- [styles](styles/README.md)
