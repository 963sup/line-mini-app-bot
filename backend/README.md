# Go 後端

## 目錄導覽與維護

cmd/ 保存 server、richmenu、firestorecheck 與 attendancemigrate 入口；internal/ 按 attendance、identity、summary、weather 等能力分工。httpapi/ 保存契約接線與產生碼。

業務核心、HTTP/LINE adapter 與 Firestore adapter 分層；共享組裝在 internal/server。修改 HTTP 格式先走 ../api 契約流程，不手改產生碼。憑證不得寫入 .env.example。

使用根目錄 scripts/tooling/tool-exec.sh；Go 修改依 ../docs/agents/acceptance.md 執行 fmt-check、lint、check、build 與相關既有測試。純文件修改不冒稱應用測試已執行。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

- [文件索引：Backend](../docs/README.md#按模組閱讀)
- [Go browser Session 與滑動續期](../docs/architecture/browser-session.md)
- [後端功能邊界](../docs/architecture/backend-boundaries.md)
- [依賴規則](../docs/architecture/dependencies.md)
- [Go 開發慣例](../docs/development/go.md)
- [可靠性與資料](../docs/architecture/reliability.md)
- [API 契約工作流程](../api/README.md)
