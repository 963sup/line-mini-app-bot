# Knowledge 設計

**Proposal** · [Design](../../proposals/knowledge.md) · [Core flow](../../proposals/knowledge.md#design-diagram) · [Acceptance](../../proposals/domain-collaboration.md) · [Module guidance](../../../backend/internal/business/knowledge/AGENTS.md)

Owns knowledge entries reusable across use cases, including sources, scope, versions and publication state. For PDF templates, callers validate field mappings; Knowledge manages confirmed versions for later queries.

[Domain guidance](../AGENTS.md)

## 後端分層結構

下列為[Knowledge 後端模組](../../../backend/internal/business/knowledge/)的完整檔案樹，包含實作、測試與目錄文件。

**範圍**　此樹列出「知識檢索」模組根目錄內的 6 個檔案。檔名以 `_test.go` 結尾的是測試；同名檔案依所在目錄區分職責。

**分層對照**　共通範本中，以下路徑未在此模組根目錄下獨立設置：`domain/entity/`、`domain/valueobject/`、`domain/service/`、`domain/event/`、`domain/policy/`、`application/command/`、`application/query/`、`application/dto/`、`adapter/http/`、`adapter/line/`、`adapter/messaging/`、`adapter/scheduler/`。子功能的分層以樹中實際路徑為準。

```text
backend/internal/business/knowledge/
├── adapter/                                  # 外部系統轉接
│   └── firestore/                            # Firestore 保存與查詢
│       ├── rules.go                          # 知識規則（讀寫 Firestore 資料）；Reader、Retrieve
│       ├── search.go                         # 規則搜尋（讀寫 Firestore 資料）；SearchEvidence
│       └── search_test.go                    # 驗證規則搜尋（讀寫 Firestore 資料）；TestSearchEvidenceChecksPublicationAndOrganization
├── application/                              # 應用用例
│   └── port/                                 # 外部能力介面
│       ├── retrieve.go                       # 知識檢索契約（定義用例所需介面）；Lookup、PublishedRule
│       └── search.go                         # 規則搜尋（定義用例所需介面）；EvidenceSearch、EvidenceSearcher
└── AGENTS.md                                 # 本目錄開發指引
```
