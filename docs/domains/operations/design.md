<a id="業務操作能力"></a>

# Business operations 設計

**Generic operation model proposed** · [Design](../../proposals/operations.md) · [Core flow](../../proposals/operations.md#design-diagram) · [Acceptance](../../proposals/domain-collaboration.md) · [Module guidance](../../../backend/internal/business/operations/AGENTS.md)

Plans reusable typed operations delegating to rule-owning domains. [PDF extraction](../../../backend/internal/business/operations/adapter/http/poppler_test.go) is implemented. [Production operations](../../operations/AGENTS.md) owns Cloud Run deployment/recovery.

[Domain guidance](../AGENTS.md)

## 後端分層結構

下列為[Business operations 後端模組](../../../backend/internal/business/operations/)的完整檔案樹，包含實作、測試與目錄文件。

**範圍**　此樹列出「文件辨識與工序推導」模組根目錄內的 10 個檔案。檔名以 `_test.go` 結尾的是測試；同名檔案依所在目錄區分職責。

**分層對照**　共通範本中，以下路徑未在此模組根目錄下獨立設置：`domain/entity/`、`domain/valueobject/`、`domain/service/`、`domain/event/`、`domain/policy/`、`application/command/`、`application/query/`、`application/port/`、`application/dto/`、`adapter/line/`、`adapter/firestore/`、`adapter/messaging/`、`adapter/scheduler/`。子功能的分層以樹中實際路徑為準。

```text
backend/internal/business/operations/
├── adapter/                                  # 外部系統轉接
│   ├── http/                                 # HTTP 入口
│   │   ├── poppler.go                        # PDF文字擷取（處理 HTTP 請求與回應）；Register、ExtractPopplerText
│   │   └── poppler_test.go                   # 驗證PDF文字擷取（處理 HTTP 請求與回應）；TestExtractionErrorResponses
│   ├── openai/
│   │   ├── candidates.go                     # 辨識候選；Extractor、ExtractCandidates
│   │   ├── candidates_test.go                # 驗證辨識候選；TestFileResponseContract
│   │   ├── procedures.go                     # 施工工序推導；SuggestProcedures
│   │   └── procedures_test.go                # 驗證施工工序推導；TestProcedureSuggestionsAreBoundedAndUntrusted
│   └── poppler/
│       └── text.go                           # 文字擷取；Extractor、Write
├── application/                              # 應用用例
│   ├── candidates.go                         # 辨識候選；Document、Candidate
│   └── document.go                           # 來源文件；DocumentTextExtractor、DocumentTextService
└── AGENTS.md                                 # 本目錄開發指引
```
