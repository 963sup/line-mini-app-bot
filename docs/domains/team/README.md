# 團隊

## 領域設計

**子域定位**　團隊治理的支援候選。

**通用語言**　管理範圍指出管理方是組織或企業；管理方識別碼指出是哪一個。群組 ID 可識別團隊，群組關係不等於管理授權。

**上下文與契約**　Group 提供群組識別，Organization 或 Enterprise 提供管理方資料；Team 同時核對管理方與團隊端資格。

**驗收案例**　團隊已有管理方時，改掛另一管理方應回傳衝突；即使操作者有目標組織管理權，也需要通過團隊端的管理者檢查。

[對照實作或測試](../../../backend/internal/business/team/firestore/store.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

團隊與組織的歸屬及管理範圍見組織所有權目標規格。

[相關業務規格](../repository/organization.md) · [後端team 實作說明](../../../backend/internal/business/team/model.go)
