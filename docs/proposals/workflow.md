# Workflow 領域設計

狀態：提案。目錄入口見[領域導覽](../../backend/internal/business/workflow/README.md)。

Workflow 定義固定步驟、等待條件及恢復位置。步驟可以呼叫 [Operations](operations.md) 執行工作，也可以委派有界任務給 [Agent](agent.md)，再按回傳結果繼續。需要跨請求恢復時才引入持久化，儲存格式與公開 API 留待使用案例決定。

## 流程模型

依 [狀態機](../references/ddd/state-machine.md)及[樂觀併發](../references/ddd/optimistic-concurrency.md)，持久化時以 `RunID` 識別一次流程，用執行版本保護轉換。候選版本、目標與命令組成確認的值，確認與提交識別碼在同一轉換保存。競爭請求的舊版本回報衝突。完成狀態只能由有效收據推進；結果未知可用原提交識別碼恢復。拒絕、過期與撤權應有明確結果。

<a id="design-diagram"></a>

## 確認與恢復流程圖

未來設計

```mermaid
flowchart TD
    extract["執行擷取操作"] --> save["保存候選版本"]
    save --> waitForUser["等待使用者確認"]
    waitForUser -->|補查後候選變更| save
    waitForUser -->|收到確認| current{"確認與目前版本相符？"}
    current -->|否| waitForUser
    current -->|是| key["保存本次寫入的執行識別碼"]
    key --> write["呼叫確認寫入操作"]
    write -->|成功並保存進度| done["流程完成"]
    write -->|進度保存前中斷| resume["恢復流程並沿用原識別碼"]
    resume --> write
```

## 流程恢復驗收

先執行擷取操作，再保存候選版本並等待使用者確認。只有確認與目前版本相符時才進入寫入步驟；補查造成候選變動時需重新確認。驗收時模擬寫入成功後、流程進度保存前發生中斷，確認恢復後以相同執行識別碼重試，只寫入一次。

## 對外契約草案

以下為邏輯簽章，實作時對照既有介面。身分及資源範圍取自可信呼叫情境。

| 操作 | 結果、副作用與失敗語意 |
|---|---|
| Start(actor, resourceID) | 建立或取回本次流程與等待步驟；記錄所引用的草稿版本。 |
| AcceptConfirmation(actor, runID, candidateVersion, target, command) | 核對確認者、流程、候選版本、目標及命令，執行前保存寫入意圖與識別碼；過期確認要求使用者重看。 |
| Resume(actor, runID) | 讀取等待點。若寫入結果未定，先查原結果，或以原碼安全重送；若版本衝突回到確認，遇權限拒絕則停止。 |

[共用架構、時序與驗收](domain-collaboration.md)

[物件呼叫文件辨識能力與驗收](domain-collaboration.md#物件呼叫文件辨識能力)
