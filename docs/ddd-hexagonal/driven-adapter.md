# 出站轉接器（Driven Adapter）

實作出站埠，執行資料庫、第三方 API 或訊息交付。

## 設計時要確認

把技術語意映射為核心契約，驗證期限、取消、交易與衝突行為。

## 示例

資料庫 Adapter 在同一工作單元內使用交易物件保存邀請及防重收據。

## 延伸閱讀

[adapter](adapter.md) · [driven-port](driven-port.md) · [unit-of-work](unit-of-work.md) · [DDD 設計索引](design-guide.md)
