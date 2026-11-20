# 議題

## 領域設計

**子域定位**　事項追蹤的支援候選。

**通用語言**　議題有建立者及開啟、關閉狀態；分類與內容編輯分別授權。

**上下文與契約**　Repository 提供資源範圍；Issue 擁有議題狀態。Project 引用議題時，專案內的操作結果不能直接取代來源議題狀態。

**驗收案例**　建立者可關閉議題後重新開啟；只有分類權限的成員不能藉由分類操作修改內容。

[對照實作或測試](../../../backend/internal/business/issue/issue_test.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

議題的建立、分類與狀態變更以議題規格為準。

[相關業務規格](../repository/issues.md) · [後端issue 實作說明](../../../backend/internal/business/issue/README.md)
