# 轉接器（Adapter）

在技術介面與 Port 契約之間轉換。

## 設計時要確認

處理協定、序列化及錯誤映射，讓模型負責業務規則。

## 示例

HTTP Adapter 將具名拒絕轉為公開回應，資料庫 Adapter 將版本前置條件失敗轉為衝突。

## 延伸閱讀

[driving-adapter](driving_adapter.md) · [driven-adapter](driven_adapter.md) · [DDD 設計索引](design_guide.md)
