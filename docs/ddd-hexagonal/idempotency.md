# 冪等（Idempotency）

同一意圖重送時只產生一次預期的邏輯效果。

## 設計時要確認

定義鍵的作用域、有效期及請求指紋；防重紀錄與受保護的寫入一起提交。

## 示例

同鍵同內容可依可見性規則回放收據，同鍵不同內容應拒絕。

## 圖解

**設計參考。**

```mermaid
flowchart TD
  R["重送請求"] --> K{"同範圍的鍵已有收據？"}
  K -->|無| W["原子提交業務寫入與收據"]
  K -->|有| F{"指紋相同？"}
  F -->|否| E["拒絕鍵被重用"]
  F -->|是| V["核對目前身分與回放可見性"]
  V -->|允許| O["回放已提交收據"]
  V -->|拒絕| E2["拒絕回放"]
```

## 延伸閱讀

[command](command.md) · [transaction](transaction.md) · [integration-event](integration-event.md) · [DDD 設計索引](design-guide.md)
