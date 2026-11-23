# Agent 設計

**Proposal** · [Design](../../proposals/agent.md) · [Core flow](../../proposals/agent.md#design-diagram) · [Acceptance](../../proposals/domain-collaboration.md) · [Module guidance](../../../backend/AGENTS.md)

Choose the next allowed operation based on task goals and current results.

[Domain guidance](../../AGENTS.md)

## 後端分層結構

下列為[Agent 後端模組](../../../backend/internal/business/agent/)的完整檔案樹，包含實作、測試與目錄文件。

**範圍**　此樹列出「證據調查」模組根目錄內的 6 個檔案。檔名以 `_test.go` 結尾的是測試；同名檔案依所在目錄區分職責。

**分層對照**　共通範本中，以下路徑未在此模組根目錄下獨立設置：`domain/entity/`、`domain/valueobject/`、`domain/service/`、`domain/event/`、`domain/policy/`、`application/command/`、`application/query/`、`application/dto/`、`adapter/http/`、`adapter/line/`、`adapter/firestore/`、`adapter/messaging/`、`adapter/scheduler/`。子功能的分層以樹中實際路徑為準。

```text
backend/internal/business/agent/
├── adapter/                                  # 外部系統轉接
│   └── genai/
│       ├── evidence.go                       # 引文證據；TextModel、EvidenceModel
│       └── evidence_test.go                  # 驗證引文證據；TestQuoteResponseContract
├── application/                              # 應用用例
│   ├── port/                                 # 外部能力介面
│   │   └── investigate.go                    # 證據調查契約（定義用例所需介面）；Investigation、Evidence
│   ├── investigator.go                       # 來源讀取與引文調查；QuoteModel、PageReader
│   └── investigator_test.go                  # 驗證來源讀取與引文調查；TestInvestigationTrustAndBudget、TestInvalidInvestigationDoesNotReadOrGenerate
└── AGENTS.md                                 # 本目錄開發指引
```
