# 驗收證據與限制

交付時記錄實際驗證的版本與環境、相關的未提交內容、命令、結果及未驗證項目。接手者依[證據重用流程](../workflow/handoff.md#minimal-handoff-and-evidence-reuse)判定可沿用的結果。

## Evidence limits

交付時明列既有測試未涵蓋的授權、重送、取消或併發情境。環境診斷與模擬資料只證明各自涵蓋的範圍。角色交接時沿用仍適用的[驗收證據](../workflow/handoff.md#minimal-handoff-and-evidence-reuse)。

Graphify AST extraction can omit TOML, documentation and dynamic calls. Repomix
is a source bundle, not a correctness proof. Configuration parsing, MCP handshake,
host tool exposure and representative tool operations are separate checks.
Do not claim a fresh Cloud environment works based on an existing tool cache.

## Cloud Run and browser location

Cloud Run changes: verify image contents, HTTP routing, service account, IAM, secrets,
revision readiness and shutdown behavior; config parsing is not a deployment test.
Geolocation changes: permission denied, unavailable, timeout, unsupported/insecure context,
coordinate validation, missing fields, cancellation/watch cleanup and authorization.
Report LINE iOS/Android and external-browser acceptance separately; mocks do not prove device support.

## Business Operation evidence

業務變更須對照 [操作契約](../../contracts/business-operations.md#功能操作契約範本) 與 [符合性流程](../workflow/implementation.md#business-operation-compliance)：

- Actor 的可信來源、Resource 歸屬、Action 與 Authority 範圍可定位至實作。
- 授權條件、領域規則、系統判定與合法轉換一致；前端不能取代後端授權。
- 依風險提供成功、越權、非法狀態及重要寫入的重送／併發證據，說明失敗不應改變的資料。
- 分別記錄 Graphify 範圍、Repomix 遺漏、Context7／官方文件來源與版本，以及實際執行、跳過或待確認項目。

測試與回歸證據依[工程決策與交付要求](../../../AGENTS.md#工程決策與交付要求)。純文件變更驗證連結與 whitespace；Codex 設定或能力指引變更另跑 check-config。

[驗收矩陣](../tooling/testing/commands.md#變更驗收矩陣) · [驗收索引](README.md)
