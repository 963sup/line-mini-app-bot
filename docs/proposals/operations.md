# Operations 領域設計

狀態：提案。目錄入口見[領域導覽](../../backend/internal/business/operations/README.md)。

資源領域可以呼叫 Operations 提供的 AI 能力。共用業務操作也可由 API、[Workflow](workflow.md) 或 [Agent](agent.md) 呼叫。每個操作定義輸入、結果及副作用，透過介面委派資源領域執行。授權沿用[業務操作模型](../architecture/contracts/business-operations.md)，由資源領域依可信呼叫者及資源範圍判定。

<a id="design-diagram"></a>

## 操作與授權流程圖

```mermaid
flowchart TD
    api["API"] --> op["具名操作"]
    flow["Workflow"] --> op
    agent["Agent"] --> op
    op --> auth{"資源領域授權通過？"}
    auth -->|否| reject["拒絕執行"]
    auth -->|是| execute["資源領域執行操作並回傳結果"]
    object["Object 應用層核對儲存庫權限"] --> extract["Operations 擷取文件候選"]
    extract --> response["模型回應與候選交回 Object"]
```

[跨領域協作與驗收](domain-collaboration.md)
