# 工項匯入

## 領域設計

**子域定位**　工程資料導入的支援候選。

**通用語言**　解析結果是候選項目；匯入執行紀錄追蹤一次操作，建立者決定該紀錄對誰可見。

**上下文與契約**　解析器提供候選資料，ObjectImport 協調匯入；Repository 的存取範圍與 Object 的工項規則仍適用。

**驗收案例**　以其他使用者的身分操作已有的自動匯入紀錄時，回傳找不到，避免揭露該筆紀錄是否存在。

[對照實作或測試](../../../backend/internal/business/objectimport/adapter/firestore/automatic.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

工項匯入的輸入、解析及執行流程見匯入規格。

[相關業務規格](../repository/object-imports.md) · [後端objectimport 實作說明](../../../backend/internal/business/objectimport/README.md)
