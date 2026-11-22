# 交易（Transaction）

具有明確提交結果及隔離語意的一組資料操作。

## 設計時要確認

說清楚哪些讀寫共用交易、哪些會被重跑，以及何時能回報完成。

## 示例

邀請狀態和收據一起提交，通知的發布安排在提交後。

## 提交與發布

**設計參考。** 本圖以 Transactional Outbox 作示例。Outbox relay 必須能重試，消費端應按事件 ID 防重。

```mermaid
sequenceDiagram
  participant App as 應用服務
  participant DB as 資料庫
  participant Relay as Outbox relay
  participant Broker as 訊息系統
  App->>DB: 同一交易保存業務狀態及 Outbox
  DB-->>App: 提交成功
  Relay->>DB: 讀取已提交的待發事件
  Relay->>Broker: 發布事件
  Broker-->>Relay: 確認接收
  Relay->>DB: 標記已發布
```

## 延伸閱讀

[consistency](consistency.md) · [unit-of-work](unit-of-work.md) · [outbox-pattern](outbox-pattern.md) · [DDD 設計索引](README.md)
