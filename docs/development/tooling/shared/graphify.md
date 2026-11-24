# Graphify 圖譜查詢與更新

關係不清楚時優先使用已曝光的 Graphify MCP；使用其他主機的圖譜時，先核對儲存庫與索引版本。AST 可能省略 TOML 或執行期關係，查得路徑後用原碼或 gopls 核對。CLI 查詢成功而 MCP 初始化逾時時，先確認本機服務回傳 JSON-RPC，再依[MCP 診斷](../../environment/troubleshooting.md#設定與-mcp-診斷)核對主機與索引。

## 安裝與設定

版本與任務以根目錄 `mise.toml` 為準。從儲存庫根目錄執行 `bash scripts/tooling/tool-exec.sh mise run install-graphify` 安裝。排除規則由 `.graphifyignore` 維護，圖譜產物位於 `graphify-out/`，不加入 Git。

## Operation

Refresh with `bash scripts/tooling/tool-exec.sh mise run graphify-update`; the task scans all code together so Go receivers in separate files resolve to their type declaration, then regenerates the report. Query with `graphify query "<question>"`.

### Docs-only graph

The existing graphify-update task indexes code only; it does not extract documentation semantics.
For a separate, credential-free Markdown navigation graph:

```bash
bash scripts/tooling/tool-exec.sh mise run graphify-docs
bash scripts/tooling/tool-exec.sh graphify query "Firestore" --graph /workspace/tooling/cache/graphify-docs/graph.json
```

The graph covers tracked and non-ignored project Markdown across the repository,
including nested module guides. Linked `.mermaid` sources appear as file nodes.
Upstream skill directories and the generated shadcn reference snapshot are
excluded. ATX headings and inline local links retain their source locations.

It does not infer semantic relationships, validate URL fragments, or automatically
identify missing requirements. Review document content to assess gaps. Generated
files stay outside the source tree and do not replace the code graph.

For synchronization, rebuild, query the affected concepts, follow related files
with path/explain, and check the source text against the accepted baseline. After
editing, rebuild and repeat the relevant queries until the scoped inconsistencies
are resolved. A valid link graph alone does not prove semantic consistency.

Full semantic extraction is a separate operation requiring a configured backend
when using the headless extract CLI. No API credentials are needed for the local
navigation graph; never store credentials in repository configuration.

Project exclusions live in `.graphifyignore` (gitignore-style). After changing
exclusions, run `graphify extract . --code-only --force --allow-dedup-shrink --out .`
through the project tooling to rescan and prune old indexed skill/example sources.
AST-only extraction does not turn Markdown plans into an implemented code graph.

## 產物更新與限制

來源刪除、搬移或排除規則改變後重建圖譜，並搜尋生成產物是否仍引用舊路徑。歷史快照與快取不代表目前來源；只清理任務擁有且確認過時的產物。社群命名若缺少選用的 `openai` 套件，可能保留 Community N 名稱；分開回報建圖與命名結果。
