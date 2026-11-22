# docs 工作指引

本目錄沿用[根目錄指引](../AGENTS.md)。文件修改遵循[文件規範](standards/documentation.md)；讀者入口見[按任務開始](README.md#按任務開始)。

## 職責與邊界

先定義讀者要完成的事、規則由誰維護，以及什麼證據可以核對，再決定文件位置。每項規則 MUST 有一個權威來源；其他文件只摘要或連結。

- 領域規則與不變條件由 `domains/` 維護。共用的身分、授權、事件與 API 語意由跨邊界契約維護；領域文件描述其具體適用方式與代表案例。
- `architecture/` 說明系統如何組成與運作。`integrations/` 說明如何與外部服務接合、設定和驗證。外部服務的能力不等於本專案已實作或已部署。
- 已採用的決策放 `decisions/`，待評估的方案放 `proposals/`。外部概念與來源可集中在 `references/`，本專案採用的規則仍由其負責分類維護。
- 開發步驟、維運操作及工程規範分別由 `development/`、`operations/` 與 `standards/` 維護。工具版本、連線設定與機密依各自權威來源管理。

## 目錄配置

以下是用途與命名範本。依實際內容建立頁面；MUST NOT 為補齊目錄樹而建立空頁、複製規則或搬移職責已清楚的文件。既有分類可保留，若需調整，應以改善查找與責任邊界為依據。

```text
docs/
├── README.md                      # 依任務查找權威文件
├── architecture/                  # 系統架構
│   ├── README.md
│   ├── overview.md                # 系統全貌與元件關係
│   ├── boundaries.md              # 模組邊界與依賴規則
│   ├── runtime.md                 # 執行流程與生命週期
│   ├── data.md                    # 資料流與一致性邊界
│   ├── security.md                # 信任邊界與安全模型
│   └── diagrams/                  # 需獨立維護的架構圖
├── domains/                       # 業務規則的權威來源
│   ├── README.md
│   └── <bounded-context>/
│       ├── README.md              # 領域概覽與邊界
│       ├── model.md               # Entity / Value Object / Aggregate
│       ├── operations.md          # Use Case / Command / Query
│       ├── policies.md            # 授權、規則與不變條件
│       ├── lifecycle.md           # 狀態轉換與生命週期
│       └── diagrams/              # 需獨立維護的領域圖
├── integrations/                  # 外部系統整合
│   ├── README.md
│   ├── line/
│   │   ├── README.md
│   │   ├── messaging.md
│   │   ├── mini-app.md
│   │   ├── configuration.md
│   │   └── rich-menu.md
│   ├── geolocation/
│   │   └── README.md
│   ├── firestore/
│   │   └── README.md
│   └── storage/
│       └── README.md
├── contracts/                     # 跨邊界契約
│   ├── README.md
│   ├── identity.md                # 身分與 Session 契約
│   ├── authorization.md           # Actor / Resource / Action
│   └── events.md                  # 跨領域事件語意
├── decisions/                     # 架構決策紀錄
│   ├── README.md
│   └── 0001-<decision>.md
├── development/                   # 開發工作流程
│   ├── README.md
│   ├── setup.md                   # 環境初始化
│   ├── workflow.md                # 開發與交付流程
│   ├── testing.md                 # 測試策略與驗收
│   ├── debugging.md               # 問題診斷
│   └── tooling.md                 # CLI / MCP / Codex
├── operations/                    # 部署與維運
│   ├── README.md
│   ├── deployment.md              # 部署流程
│   ├── configuration.md           # 執行環境設定
│   ├── observability.md           # Logs / Metrics / Traces
│   ├── troubleshooting.md         # 故障排查
│   └── recovery.md                # 復原與回滾
├── standards/                     # 工程規範
│   ├── README.md
│   ├── engineering.md             # 開發與設計原則
│   ├── naming.md                  # 命名規則
│   ├── documentation.md           # 文件維護規範
│   └── quality.md                 # 品質與驗收門檻
├── proposals/                     # 尚未採用的方案
│   ├── README.md
│   └── <proposal>.md
└── references/                    # 外部知識與技術參考
    ├── README.md
    └── ddd.md                     # DDD 概念與參考來源
```

目前的跨領域契約、部署指南與 DDD 索引分別從 [跨邊界契約](contracts/README.md)、[部署與維運](operations/README.md) 與 [DDD 參考](references/ddd/README.md) 查閱。

## 文件與目錄命名規範

文件與目錄 MUST 依實際職責及業務語意命名，遵循高內聚、單一責任、簡潔明確及一致性原則。一般目錄與文件名稱統一使用小寫英文 `kebab-case`，優先採用標準領域術語，避免模糊名稱、重複上層目錄語意、冗餘前綴及非必要縮寫。目錄入口固定使用 `README.md`，Agent 指引使用 `AGENTS.md`，架構決策使用四位數編號加 `kebab-case`。既有名稱僅在能改善語意或責任邊界時調整，並同步更新所有引用與連結；不得為命名一致性進行無實質價值的重構。

<a id="主題子目錄"></a>

## 文件維護

`README.md` 引導讀者選取主題，`AGENTS.md` 補充上層未涵蓋的修改規則。新增或搬移頁面時更新所屬分類索引及直接引用；入口變更時才更新上層導覽。區分現行實作、待驗收行為與未來設計，歷史變更由 Git 查閱。

## 圖表維護

圖表及其實作或驗收證據 MUST 由主題頁維護，依[文件規範](standards/documentation.md#圖表狀態與入口)標明現行實作、未來設計、開發與維運指引或現況與提案對照。新增或變更圖表時，MUST 更新直接引用與受影響的分類入口，並保留仍被引用的錨點。只在[專案閱讀入口](README.md#日常開發看哪張圖)選列跨主題常用圖。

## 驗證

核對相對路徑、來源、標題錨點與索引。改圖時依[圖表規範](standards/documentation.md)驗證圖意與渲染；命令與證據依[驗收矩陣](development/acceptance.md#變更驗收矩陣)。
