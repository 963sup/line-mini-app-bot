# shadcn 本機參考庫

## 用途與範圍

保存 shadcn 官方文件的本機參考入口。

## 架構與責任邊界

快照與索引依既有流程產生，專案的元件規則由前端 UI 指引維護。

## 閱讀與修改入口

先查閱[snapshot/INDEX.md](snapshot/INDEX.md)。

## 相關規格

[生成索引](snapshot/INDEX.md) · [MIT 授權](LICENSE.md)


此處保存官方 [Components](https://ui.shadcn.com/docs/components) 與 [AI Elements](https://ui.shadcn.com/docs/ai-elements) 入口列出的全部項目文件，供 Codex 搜尋選用。完整清單、用途、Base UI／Radix 連結及棄用提醒只由 [生成索引](snapshot/INDEX.md) 維護；不再手工複製元件目錄。

## 搜尋與選用

UI 任務先搜尋完整索引，再讀相關 Radix 文件及內嵌範例，最後以固定 CLI／registry 核對實際安裝內容。元件保護、主動選用與驗收遵循 [UI 指引](../../../../frontend/miniapp/src/shared/ui/AGENTS.md#需求驅動的選用與驗收)。

從儲存庫根目錄執行：

```bash
rg -n -i 'select|search|dialog|chat|reasoning' docs/development/tooling/shadcn-reference/snapshot/INDEX.md
rg -n '## |```tsx|Installation' docs/development/tooling/shadcn-reference/snapshot/pages/ai-elements/radix/chat.md
bash scripts/tooling/tool-exec.sh shadcn info --json --cwd frontend/miniapp
```

不要將整份參考庫注入每次任務。它不代表已安裝元件；文件中的別名、路由、`@latest`、模型服務及安裝命令都只是上游範例，不是專案規則或變更授權。依本專案的 base、aliases、Pages Router 及固定 CLI 調整用法。

## 更新與完整性

更新由 [下載腳本](../../../../scripts/tooling/shadcn-reference.py) 負責，不需要額外 Python 套件：

```bash
# 下載當前官方目錄及每項的 Base UI／Radix Markdown；全部驗證後才替換舊快照。
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/shadcn-reference.py

# 離線核對目錄覆蓋、檔案雜湊、來源、base 標記、索引與七日新鮮度。
bash scripts/tooling/tool-exec.sh python3 scripts/tooling/shadcn-reference.py --check
```

開始 UI 任務時先執行離線檢查；缺少、損壞或超過七日就主動更新，再搜尋索引。同次任務且快照未變時重用結果。更新失敗會保留舊快照，須明列其日期與受阻原因；舊資料只供探索，受影響的實作決策仍須取得當前官方文件／registry 證據。

[manifest.json](snapshot/manifest.json) 記錄每份文件的請求／最終網址、抓取時間、HTTP ETag／Last-Modified（若有）、SHA-256、官方描述、base 及棄用標記。索引來自當次官方入口，不固定寫死數量；更新途中若入口清單改變就停止，避免混入不完整目錄。HTTP 快照不是不可變 Git revision，也不保證所有頁面屬於同一部署版本；不宣稱與固定 CLI 同版。

## 內容邊界與授權

`snapshot/pages/` 保留上游 Markdown 原文及其已內嵌的程式範例，不手工修改。未遞迴下載外部 API、圖片、影片、示範專案、所有依賴或 `<ComponentSource>` 等未展開來源；需要實際元件原始碼時仍用固定 CLI 預覽。這是兩個官方入口列出的完整文件範圍，不是所有第三方 registry 或整個 shadcn 網站的鏡像。

同時保存兩種 base 供辨識差異，本專案仍優先使用 Radix。若官方網址和 frontmatter 的 base 不同，或文件已棄用，索引會標示，不把它冒充可直接安裝的 Radix 實作；例如 Radix Toast 文件要求改用 Sonner。

上游來源為 shadcn/ui，保留其 [MIT 授權](LICENSE.md)，與當次讀取的 [官方授權檔](https://github.com/shadcn-ui/ui/blob/main/LICENSE.md) 逐位元組核對；文件原文來源見 manifest。生成快照不套用自有 Markdown 的標題／連結重寫規範，其相對網址仍依官方網站解析。一般 Graphify／Repomix 分析排除生成快照，使用 `rg` 或直接讀取相關文件；參考庫說明、下載腳本與專案規則仍納入分析。

## 本次查證

2026-11-18（Asia/Taipei）：直接讀取官方 Markdown 入口及各項文件，覆蓋與雜湊由腳本檢查。Graphify 用於定位工具關係，其 AST 圖不涵蓋全部文件規則；Repomix 用於專案指引交叉核對。Context7 本次 resolve 成功，但 query 達月額度，未宣稱取得新結果；CLI 預覽流程沿用本聊天先前已取得的 Context7 證據，下載格式則以當次官方回應驗證。未安裝應用元件或執行 UI／LINE 驗收。

實際通過：下載及離線完整性檢查、`check-config.py`、維護文件的本機連結／錨點、`git diff --check`；以暫存副本確認文件損壞與過期會被拒絕、下載失敗保留舊快照。Repomix 交叉核對八個相關檔案；Graphify 已執行 AST 更新，未執行 LLM 語意抽取。分析範圍檢查通過，但其 Git 清單不包含尚未追蹤的新檔，不將該結果冒充新快照的覆蓋證據。

[工具索引](../README.md) · [shadcn 工具指南](../mcp/shadcn.md) · [文件總索引](../../../README.md)

[開發與維運指引圖索引](../../../diagrams.md#開發與維運指引) · [未來設計圖](../../../diagrams.md#未來設計)。
