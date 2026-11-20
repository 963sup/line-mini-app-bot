# 群聊摘要

## 領域設計

**子域定位**　群聊資訊整理的支援候選。

**通用語言**　來源訊息是可引用的證據；摘要是附來源編號的生成結果；訊息編輯或刪除會影響可用證據。

**上下文與契約**　LINE 提供訊息事件，模型供應商提供候選文字。Summary 擁有收錄及輸出驗證規則，決定結果能否儲存或發送。

**驗收案例**　更換模型後仍須檢查引用；結果缺少引用或引用超出來源範圍時，拒絕儲存與發送。

[對照實作或測試](../../../backend/internal/business/summary/summarization/validation.go) · [戰略設計與分類準則](../../architecture/ddd/README.md)

摘要規格涵蓋訊息收錄、產生摘要及來源引用。

[相關業務規格](summary.md) · [後端summary 實作說明](../../../backend/internal/business/summary/README.md)
