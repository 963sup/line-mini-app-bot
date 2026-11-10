# Graphify／Repomix 交叉比對與巢狀指引

日期：2026-11-11（Asia/Taipei）。分析基準提交：`cc82d1ef8c3d8baa58216a5e029bc65904a3b661`，包含本次尚未提交的巢狀文件。已 fetch 並核對 origin/work 與 origin/main。Git 發布結果以實際推送後的遠端查核為準。

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
