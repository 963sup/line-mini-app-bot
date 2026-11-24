# 查詢（Query）

取得資訊且不改變業務狀態的用例。

## 設計時要確認

定義可見性、分頁與資料鮮度；讀模型可直接投影所需欄位。

## 示例

查詢邀請列表可讀投影，若投影落後則以契約揭露時效。

## 延伸閱讀

[cqrs](cqrs.md) · [eventual-consistency](eventual-consistency.md) · [DDD 設計索引](design-guide.md)
