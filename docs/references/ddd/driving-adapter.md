# 入站轉接器（Driving Adapter）

接收外部請求並呼叫入站埠。

## 設計時要確認

解析輸入、確認身分與映射結果；業務資格由用例與模型處理。

## 示例

HTTP Handler 從經驗證的身分建立用例輸入，避免採信 payload 宣告的角色。

## 延伸閱讀

[adapter](adapter.md) · [driving-port](driving-port.md) · [DDD 設計索引](README.md)
