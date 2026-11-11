# Acceptance and evidence

Current phase: stabilize the foundation; do not add test plans, cases or tools.
The test-related matrix entries below are retained for the later testing phase.
Keep existing CI; this phase note does not authorize removing checks or claiming
unrun checks passed. Scope follows the [accepted baseline](../architecture/decisions/0002-baseline.md).



| Change | Required evidence |
| --- | --- |
| Configuration/scripts | `python3 scripts/tooling/check-config.py`; relevant task smoke check |
| Go behavior | format check, focused package tests, then shared `mise run check` and `mise run build` before delivery |
| Concurrency | race tests exercising the changed concurrent behavior |
| Dependencies | inspect go.mod/go.sum; download/verify modules; vulnerability review as appropriate |
| LINE authentication/webhook | invalid signature/token, unauthorized user and duplicate event cases |

Run tooling through `bash scripts/tooling/tool-exec.sh <command>`. Formatting can be
checked with `gofmt -l` on changed Go files; formatting changes require file
ownership. No go.mod means application verification is unavailable, not passed.
Run complete race suites once after integration instead of concurrently per role.

## Evidence limits

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

業務變更須對照 [操作契約](../architecture/business-operation.md#功能操作契約範本) 與 [符合性流程](workflow.md#business-operation-compliance)：

- Actor 的可信來源、Resource 歸屬、Action 與 Authority 範圍可定位至實作。
- 授權條件、領域規則、系統判定與合法轉換一致；前端不能取代後端授權。
- 依風險提供成功、越權、非法狀態及重要寫入的重送／併發證據，說明失敗不應改變的資料。
- 分別記錄 Graphify 範圍、Repomix 遺漏、Context7／官方文件來源與版本，以及實際執行、跳過或待確認項目。

本階段沿用既有 CI 與適用檢查，不因本規範自動新增測試方案、案例或工具；tester 的一般職責受此階段限制。需要新增案例時先明列缺口與所需階段調整，不用缺少證據宣稱通過。純文件變更驗證連結與 whitespace；角色設定變更另跑 check-config。
