# Graphify／Repomix 交叉比對與巢狀指引

日期：2026-11-11（Asia/Taipei）。最新文件導覽整理基準提交：`b417885b39d1ac89484a56e0494b13f92126e5c1`，包含本次文件修改。下方「本次文件導覽整理」為最新結果；其後保留先前巢狀指引建立時的執行紀錄。

## 巢狀分類補齊（2026-11-11）

在五個模組閱讀入口之外，補齊 docs 七個主題目錄的 README.md／AGENTS.md：architecture、features、development、integrations、operations、tooling、agents。
各層 README 列出該層全部直接文件與既有下層索引；AGENTS.md 說明分類邊界、父層規則、連結同步與驗證要求。
更新 backend、api、frontend、deploy 的巢狀指引，新增 scripts/AGENTS.md，並由根 README 與 docs 總索引連入。
既有 decisions、tooling/mcp 索引及 frontend/miniapp 上游指引保留。

本次輸出目錄：`/workspace/tooling/reports/docs-nested-review-20261111/`。

```bash
bash scripts/tooling/tool-exec.sh graphify extract . --code-only --out /workspace/tooling/reports/docs-nested-review-20261111
bash scripts/tooling/tool-exec.sh repomix --style xml --parsable-style --output /workspace/tooling/reports/docs-nested-review-20261111/repomix-final.xml
python3 scripts/analysis/docs-graph.py --out /workspace/tooling/reports/docs-nested-review-20261111/docs-graph
```

乾淨 AST 圖譜、最終打包、涵蓋差異與結構完整性以同目錄 crosscheck-final.json、repomix-final.log 為準。
直接檢查每個主題的文件索引覆蓋與本機連結／標題錨點，補足 Graphify AST 不解析 Markdown 的限制。
本次僅整理本機文件，未更動程式、雲端設定或上游技能；其他工作中的 Dockerfile、部署文件與根指引修改均保留。
後續章節為前次整理的快照，token 數不代表本次最終打包。

## 本次文件導覽整理

保留 docs 的 architecture、features、development、integrations、operations、tooling、agents 主題分類，
在 [文件索引](../README.md#按模組閱讀) 建立 Backend／API／Frontend／Scripts／Deploy 閱讀入口。
根 README 導向此入口，各模組 README 加入相關文件及返回索引的連結；docs/AGENTS.md 同步規範文件歸屬。
跨模組功能與契約維持單一來源，未搬移既有文件。

重新執行 Graphify 0.9.103、Repomix 1.18.1：

```bash
bash scripts/tooling/tool-exec.sh graphify extract . --code-only --out /workspace/tooling/reports/docs-routing-20261111
bash scripts/tooling/tool-exec.sh repomix --style xml --parsable-style --output /workspace/tooling/reports/docs-routing-20261111/repomix-before.xml
bash scripts/tooling/tool-exec.sh repomix --style xml --parsable-style --output /workspace/tooling/reports/docs-routing-20261111/repomix-final.xml
```

- 乾淨 Graphify：1,226 節點、3,513 關係、165 個來源檔。
- 修改前 Repomix：282 檔案、302,730 tokens；修改後數值見本機 `repomix-final.log`，避免將報告自身更新造成的 token 變化混為同一快照。
- 165 個 Graphify 來源全數收錄，重複 ID、缺失關係端點及不存在來源均為 0。
- AST 不涵蓋 docs Markdown；另以既有 `scripts/analysis/docs-graph.py` 建立標題與連結圖，並直接驗證本次修改文件的本機路徑與標題錨點。
- `.agents`、`.vscode` 未被一般打包收錄，本次未修改這兩處；未把工具排除視為空目錄。
- 本次為文件導覽修改，依本機文件、原始碼目錄與 CLI 結果判斷，未變更套件或外部 API 行為，無新增外部技術選型。
- 驗證包含文件連結、設定檢查、分析範圍分類與 whitespace；未執行應用測試或雲端驗收。

完整本機產物位於 `/workspace/tooling/reports/docs-routing-20261111/`，不納入 Git。

## 執行方式與結果

透過 `scripts/tooling/tool-exec.sh` 執行 mise.toml 所固定的 Graphify 0.9.103 與 Repomix 1.18.1：

```bash
graphify extract . --code-only --out /workspace/tooling/reports/nested-docs-publish-20261111
repomix --style xml --parsable-style --output /workspace/tooling/reports/nested-docs-publish-20261111/repomix.xml
```

Graphify 使用全新輸出目錄，驗證後同步至本機 graphify-out；避免增量合併保留搬移前的 cloudbuild.yaml、go.mod。圖譜及 XML 為本機忽略產物，不納入 Git。

- Graphify：**1,226 節點、3,513 關係、165 個來源檔**。
- Repomix：**282 檔案**；完整 tokens 與安全掃描結果見同目錄 repomix.log。數值對應本報告改寫前的輸入快照，報告改寫後另輸出 repomix-final.xml／repomix-final.log。
- Graphify 來源全部包含於 Repomix；額外 117 個打包檔主要為文件及設定。
- 重複節點 ID、缺失關係端點、不存在的來源檔均為 0。

| 目錄 | Graphify 來源檔 | Repomix 檔案 |
| --- | --- | --- |
| `.agents/` | 0 | 0 |
| `.codex/` | 0 | 13 |
| `.github/` | 1 | 7 |
| `.vscode/` | 0 | 0 |
| `api/` | 7 | 13 |
| `assets/` | 0 | 3 |
| `backend/` | 106 | 111 |
| `deploy/` | 1 | 4 |
| `docs/` | 0 | 58 |
| `frontend/` | 29 | 35 |

## 更新與維護決策

根目錄及十個指定目錄均有 AGENTS.md／README.md，提供職責、修改邊界、驗證方式與父層導覽。保留既有 API 契約文件、frontend/miniapp 的 Next.js 指引與上游技能；.agents 角色數以配置為準。

`.gitignore` 只為 `.agents/AGENTS.md` 加入例外，讓專案指引可隨 Git 保存。Graphify／Repomix 仍依各自 ignore 規則排除 .agents；.vscode 本次也未收錄，已直接檢查兩處文件與 JSON，不將未收錄解釋為空目錄。

## 驗證與限制

執行設定檢查、分析範圍分類、文件本機連結檢查與 git diff --check。此變更僅為文件及 Git 忽略例外，不更改應用行為，未重跑應用測試。根 AGENTS.md 原有刪除另留在工作樹，不納入本次提交。

AST 不保證涵蓋文件、設定、圖片與動態行為；涵蓋率與安全掃描不證明語意正確或絕無機密。未宣稱官方 API 文件查證、MCP 主機驗收或雲端部署通過。main 推送可能觸發既有 CI/CD，其結果需獨立查核。
