# 工具與分析

保存工具鏈、版本管理、分析與 MCP 入口。

工具使用方式放在此處，版本和可執行任務由 mise 維護。

遇到具體工具需求再讀對應頁面；工具已安裝、已設定、可連線與操作成功是不同證據。

## 工具文件目錄標準

工具指南依主要用途歸類。通用操作放在 `shared/`，代理／MCP 頁引用它；領域專屬工具保留在前端、後端、API 或測試分類，不因有 MCP 就搬進 `agents/`。Agent Skills 與文件查詢 MCP 的主體用途屬代理，留在 `agents/`。

```text
docs/development/tooling/
├── README.md
├── shared/                 # 跨領域共用工具
├── frontend/               # 前端專屬工具
│   ├── README.md
│   ├── shadcn.md
│   └── nextjs-devtools.md
├── backend/                # 後端專屬工具
├── api/                    # API 契約、產生與驗證工具
├── testing/                # 測試與品質工具
└── agents/                 # AI Agent 與 MCP 工具
```

## 日常工具入口

| 需求 | 文件 |
| --- | --- |
| 查閱 GitHub 儲存庫、PR、CI 與 Wiki | [GitHub](shared/github.md) |
| 修復 GitHub Actions／PR 紅燈 | [檢查失敗處理流程](shared/github.md#github-actionspr-檢查失敗) |
| 使用 gcloud、登入與 ADC | [Google Cloud CLI](shared/gcloud.md) |
| 渲染與驗收 Mermaid 圖表 | [Mermaid](shared/mermaid.md) |
| 初始化共用工具鏈與 MCP | [初始化模式與操作](shared/toolchain-setup.md) |
| 工具版本、任務與環境 | [mise](shared/mise.md)；版本與任務以 [mise.toml](../../../mise.toml)為準 |
| Go 工具準備與開發操作 | [Go 開發](backend/go-development.md) |
| Go 語言工具 | [gopls](backend/gopls.md) |
| 修改與驗證 HTTP 契約 | [API 工具流程](api/change-api.md) |
| LINE Rich Menu 圖片版面 | [格線與按鈕](frontend/rich-menu-layout.md) |
| 產生 LINE Rich Menu 圖片與熱區 | [共用定義與產生器](shared/rich-menu-generator.md) |
| Firestore 工具 | [後端工具與驗證](backend/firestore.md) |
| 選擇測試與品質檢查 | [驗收矩陣與命令](testing/commands.md) |
| 使用 Codex CLI 開啟專案或提交 Cloud 任務 | [Codex](agents/codex.md) |
| 本機容器建置與 smoke 驗收 | [容器驗收](testing/container.md) |
| 選擇或診斷 MCP | [MCP 索引](agents/README.md) |
| 使用專案技能 | [Skills](agents/skills.md) |
| 查詢程式關係 | [Graphify](shared/graphify.md) |
| 打包限定原始碼 | [Repomix](shared/repomix.md) |
| 產生本次交談的來源參考技能 | [Repomix 參考技能](shared/repomix.md#產生參考技能) |
| 查 UI 元件與整合流程 | [前端工具入口](frontend/README.md) |
| 檢查 Next.js 開發執行狀態 | [Next DevTools](frontend/nextjs-devtools.md) |

## 上游參考資料

`frontend/shadcn-reference/snapshot/` 是上游參考快照，按索引搜尋所需頁面；不作為專案政策，也不全量載入。[查詢與更新方式](frontend/shadcn-reference/README.md)

分析產物存放於任務暫存目錄，核對來源版本、實際涵蓋檔案及有效性；不在 docs 累積日期報告。

[本目錄指引](AGENTS.md) · [父層指引](../../AGENTS.md) · [文件總入口](../../README.md)

[上層導覽](../README.md)
