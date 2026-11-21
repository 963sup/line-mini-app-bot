# 架構測試（Architecture Testing）

把依賴與邊界約束轉為可自動檢查的規則。

## 設計時要確認

對 Go import 圖檢查 domain 是否依賴 SDK 和 Adapter，對埠另做契約測試；例外須具名並有理由。

## 示例

一個測試抓 domain 引入 Firestore，另一個以真正儲存驗證版本衝突。

## 延伸閱讀

[dependency-rule](dependency-rule.md) · [domain-testing](domain-testing.md) · [driven-adapter](driven-adapter.md) · [DDD 設計索引](README.md)
