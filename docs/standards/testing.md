# 測試規範

驗證依變更風險選擇，重點是證明受影響行為，而不是只讓命令成功。

## 驗證矩陣

| 變更          | 最低驗證                                                 |
| ------------- | -------------------------------------------------------- |
| Markdown      | 相對連結、舊路徑搜尋、`git diff --check`                 |
| Go            | 受影響套件測試、`go -C backend vet ./...`、必要時建置    |
| 前端          | `pnpm --filter @repo/miniapp check` 與相關測試           |
| OpenAPI       | 契約解析、產生碼差異、Go 與 TypeScript 消費端            |
| Firestore     | 授權、讀寫集合、交易重試、重送、版本衝突與 emulator 案例 |
| LINE／Session | 無效簽章或 token、越權、CSRF、重送與真實裝置缺口         |
| 部署          | 映像、revision、設定、健康檢查、核心 smoke、觀察與回退   |

## 測試層級

- Domain 單元測試驗證純規則、狀態轉換與拒絕案例。
- Application 測試驗證用例協調、port 呼叫與失敗語意。
- Adapter 整合測試驗證序列化、資料庫交易及供應端轉譯。
- 契約測試驗證 OpenAPI 與產生型別、路由和消費端一致。
- 瀏覽器或裝置測試驗證使用者可觀察流程；不能由 mock 取代。

## 證據規則

測試必須確認案例實際執行；套件成功但案例被 skip、filter 沒選中或只命中快取，不算該行為的證據。失敗或環境阻塞時，記錄命令、失敗點與尚未驗證的行為，不宣稱整合通過。

Firestore emulator 不證明正式 IAM、索引、TTL 或負載；本機 HTTP smoke 不證明 LINE 真實渠道與裝置；成功建置不證明部署成功。

## 驗證責任

- MUST select verification methods according to the impact and risk of each change.
- MUST execute applicable formatting, linting, type, architecture, contract, and regression checks.
- MUST verify changed behavior against the original requirements and acceptance criteria.
- MUST verify applicable authorized, denied, invalid, and failure scenarios.
- MUST inspect generated changes and affected dependencies.
- MUST record actual verification results and material limitations.

- MUST NOT disable, bypass, or weaken required tests, lint rules, architecture checks, or CI checks.
- MUST NOT modify expected test results merely to match incorrect implementation behavior.
- MUST NOT use mocks or stubs to claim verification of untested integration behavior.
- MUST NOT treat stale generated artifacts as authoritative evidence.
- MUST NOT claim that an unexecuted check passed.
- MUST NOT substitute superficial validation for applicable behavioral verification.

- MAY introduce benchmarks for measurable performance requirements.
- MAY introduce property-based tests for complex invariants.
- MAY introduce fuzz testing for relevant input boundaries.
- MAY use emulator-based integration tests for local service verification.
- MAY introduce end-to-end tests for critical cross-system flows.
- MAY expand verification beyond the minimum required checks when risk justifies the additional cost.

## 驗證與效益

- SHOULD automate recurring mechanical verification.
- SHOULD prioritize regression checks for high-risk business behavior.
- SHOULD establish measurable baselines before optimization.
- SHOULD compare results before and after material changes.
- SHOULD prioritize removing complexity before introducing optimization mechanisms.
- SHOULD stop further optimization when additional changes lack demonstrated value.

- SHOULD NOT run unrelated expensive checks when narrower verification provides sufficient evidence.
- SHOULD NOT rely only on snapshots, type checks, or mocks for behavior requiring integration evidence.
- SHOULD NOT optimize without a measurable baseline and acceptance target.
- SHOULD NOT retain an optimization whose complexity exceeds its demonstrated value.
- SHOULD NOT expand a fix into unrelated cleanup without authorization or a shared root cause.
