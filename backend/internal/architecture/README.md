# backend/internal/architecture 目錄導覽

以架構測試檢查 Go 套件的依賴方向與分層邊界。

依目錄用途提供後端入口或共用能力；服務接線由 bootstrap 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

從 [boundaries_test.go](boundaries_test.go) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

測試涵蓋所有業務 `domain`／`application` 子套件的 SDK、外層依賴及 JSON／Firestore 格式標記，包含尚未列入具名依賴清單的新套件；HTTP adapter 子套件也不得直接引用儲存或模型供應商 SDK。依據為 [DDD 依賴規則](../../../docs/references/ddd/dependency-rule.md)。

格式相容性由 adapter 驗證：[日誌保存格式與防重雜湊](../business/journal/adapter/firestore/payload_test.go)、[組織 HTTP 回應](../business/organization/adapter/http/responses_test.go)及[模型候選資料](../business/operations/adapter/openai/candidates_test.go)。架構測試不取代這些契約與整合案例。

## 閱讀入口

[現行實作圖：格式與資料相容](../../../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

[上層導覽](../README.md)
