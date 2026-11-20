# 子工項

## 領域設計

**子域定位**　工程工項細分的支援候選。

**通用語言**　父工項界定子工項的歸屬；父工項建立者是部分寫入操作核對的身分。

**上下文與契約**　Subobject 依賴 Repository 與父 Object 的資料，處理子工項操作；日誌可引用子工項，但不能經由引用變更其父工項。

**驗收案例**　在要求父工項建立者資格的操作中，即使操作者可讀取該儲存庫，也不能以他人身分修改子工項。

[對照實作或測試](../../../backend/internal/business/subobject/adapter/firestore/store.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

子工項依附父工項；父子關係與操作範圍可對照工項規格及子工項實作。

[相關業務規格](../repository/objects.md) · [後端subobject 實作說明](../../../backend/internal/business/subobject/README.md)
