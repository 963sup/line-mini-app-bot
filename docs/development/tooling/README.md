# 工具與分析

## 用途與範圍

保存工具鏈、版本管理、分析與 MCP 入口。

## 架構與責任邊界

工具使用方式放在此處，版本和可執行任務由 mise 維護。

## 閱讀與修改入口

先查閱[mise.md](mise.md)、[analysis-scope.md](analysis-scope.md)、[mcp](mcp/README.md)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[mise](mise.md) · [MCP 索引](mcp/README.md)


遇到具體工具需求再讀對應頁面；工具已安裝、已設定、可連線與操作成功是不同證據。

## 日常工具入口

| 需求 | 文件 |
| --- | --- |
| 工具版本、任務與環境 | [mise](mise.md)；版本與任務以 [mise.toml](../../../mise.toml)為準 |
| 選擇或診斷 MCP | [MCP 索引](mcp/README.md) |
| 使用專案技能 | [Skills](skills.md) |
| 理解分析涵蓋與限制 | [分析範圍](analysis-scope.md) |
| 查詢程式關係 | [Graphify](mcp/graphify.md) |
| 打包限定原始碼 | [Repomix](mcp/repomix.md) |
| 查 UI 元件與整合流程 | [shadcn 指南](mcp/shadcn.md)、[本機參考庫入口](shadcn-reference/README.md) |
| 檢查 Next.js 開發執行狀態 | [Next DevTools](mcp/next-devtools.md) |

## 上游參考資料

`shadcn-reference/snapshot/` 是上游參考快照，按索引搜尋所需頁面；不作為專案政策，也不全量載入。[查詢與更新方式](shadcn-reference/README.md)

分析產物存放於任務暫存目錄，依[來源與證據規則](analysis-scope.md)核對版本及有效性；不在 docs 累積日期報告。

[本目錄指引](AGENTS.md) · [父層指引](../../AGENTS.md) · [文件總入口](../../README.md)

[開發與維運指引圖索引](../../diagrams.md#開發與維運指引) · [未來設計圖](../../diagrams.md#未來設計)。
