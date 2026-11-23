# Workflow 設計

**Proposal** · [Design](../../proposals/workflow.md) · [Core flow](../../proposals/workflow.md#design-diagram) · [Acceptance](../../proposals/domain-collaboration.md) · [Module guidance](../../../backend/internal/business/workflow/AGENTS.md)

Manages explicit step ordering, waiting conditions and progress.

[Domain guidance](../AGENTS.md)

## 後端分層結構

下列為[Workflow 後端模組](../../../backend/internal/business/workflow/)的完整檔案樹，包含實作、測試與目錄文件。

**範圍**　此樹列出「工作流程」模組根目錄內的 5 個檔案。檔名以 `_test.go` 結尾的是測試；同名檔案依所在目錄區分職責。

**分層對照**　共通範本中，以下路徑未在此模組根目錄下獨立設置：`domain/entity/`、`domain/valueobject/`、`domain/service/`、`domain/event/`、`domain/policy/`、`application/command/`、`application/query/`、`application/port/`、`application/dto/`、`adapter/http/`、`adapter/line/`、`adapter/firestore/`、`adapter/messaging/`、`adapter/scheduler/`。子功能的分層以樹中實際路徑為準。

```text
backend/internal/business/workflow/
├── application/                              # 應用用例
│   ├── confirmation.go                       # 操作確認；Candidate、Run
│   ├── confirmation_test.go                  # 驗證操作確認；TestConfirmedWriteCanResumeAfterProgressSaveFails
│   ├── prepare.go                            # 草稿準備；Scope、Preparation
│   └── prepare_test.go                       # 驗證草稿準備；TestPreparationComposesOptionalCapabilities
└── AGENTS.md                                 # 本目錄開發指引
```
