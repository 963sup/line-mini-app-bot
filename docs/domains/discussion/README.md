# 討論

## 領域設計

**子域定位**　協作支援候選。

**通用語言**　討論串承載對話；留言是串內回應；分類描述討論的歸類。

**上下文與契約**　Repository 提供討論所屬範圍。Discussion 決定更新討論內容與更新分類各自需要的資格。

**驗收案例**　只具分類權限的成員可以調整分類；同一請求若夾帶內容修改，應拒絕越權變更。

[對照實作或測試](../../../backend/internal/business/discussion/classification_test.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

討論串、留言與分類的行為由討論規格定義。

[相關業務規格](../repository/discussions.md) · [後端discussion 實作說明](../../../backend/internal/business/discussion/README.md)
