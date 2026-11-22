# 共用實作

## 領域設計

**子域定位**　技術共用目錄，無獨立子域分類。

**通用語言**　orghttp 負責驗證入口、解碼與回應映射；orgstore 提供 Channel 範圍的資料讀取與交易工具。

**上下文與契約**　Enterprise、Organization 與 Team 使用這些實作。共同使用交易函式並不足以宣告 DDD 的 Shared Kernel；所有權與移轉資格仍由所屬業務維護。

**驗收案例**　操作回傳歸屬衝突時，HTTP 映射為 409；請求未通過身分驗證時不得呼叫業務操作。

[對照實作或測試](../../../backend/internal/business/internal/adapter/http/response.go) · [戰略設計與分類準則](../../references/ddd/README.md)

對應 Go 業務模組共用的 orghttp 與 orgstore，供查閱組織相關的 HTTP 與儲存實作。

[後端orghttp 實作說明](../../../backend/internal/business/internal/adapter/http/response.go) · [後端orgstore 實作說明](../../../backend/internal/business/internal/adapter/firestore/store.go)

[適用工作指引](../AGENTS.md) · [上層導覽](../README.md)
