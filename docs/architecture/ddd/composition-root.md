# 組裝根（Composition Root）

程式入口附近建立與連接實作的地方。

## 設計時要確認

集中處理設定、資源生命週期及注入；業務判斷由模型與用例承擔。

## 示例

啟動時建立資料庫 client、Repository 和用例，再把用例交給 HTTP Handler。

## 延伸閱讀

[dependency-inversion](dependency-inversion.md) · [hexagonal-architecture](hexagonal-architecture.md) · [DDD 設計索引](README.md)
