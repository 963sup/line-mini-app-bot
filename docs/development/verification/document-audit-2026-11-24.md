# 文件全量盤點與未完成整改矩陣

> 狀態：部分完成／受阻。本頁是 2026-11-24 工作樹快照的續接證據；`B` 不是通過，不能據此宣稱文件體系已完成驗收。

## 基準與方法

- 範圍：`docs/**` 的 363 個現有檔案、120,940 行；含 115 份自有非 DDD Markdown、69 份唯讀 DDD 及 179 份上游 shadcn 參考檔。
- 建基準時工作樹已有 52 個受追蹤的 `docs/**` 未提交差異及 1 個未追蹤 DDD README，另有範圍外差異。新增整改均以當時工作樹為基準；既有差異不歸因於本次盤點。
- 在重新掃描時，`docs/development/tooling/shared/mise.md` 與 `docs/standards/documentation.md` 的內容雜湊相對本次基準已改變，並非本輪整改寫入；因此語意對照須在工作樹穩定後重做。
- Repomix 1.18.1 使用 `--include docs/**` 打包 185 份、275,640 tokens；上游快照依 `.repomixignore` 排除，另外以檔案系統逐檔統計。打包檔位於 `/tmp/line-mini-app-bot-docs-baseline.xml`，不進入 Git。
- `pnpm docs:check`：129 份自有 Markdown 格式通過，markdownlint 0 issue。`python3 scripts/check/check-config.py`：201 份文件的既有本地連結與錨點檢查通過。自有文件逐檔掃描未見無語言 fenced block 或標題跳級。
- 專案鎖定 Mermaid CLI 13.0.0-alpha.3 對 90 段圖表渲染全部成功。這只驗證語法，不驗證業務流程或依賴方向。
- 精確段落重複掃描只發現 4 對跨頁共用引用／相同句子，未證明存在重複權威規則。語意重複、矛盾與實作偏差仍須逐條核對。

## 數量與已確認問題

- 整改前：363 個檔案、120,940 行；整改後：364 個檔案、121,372 行。新增的矩陣為 432 行；維運設定頁的行數不變。
- 已確認的文件呈現問題 2 處，均為整段文字被包成單一連結，已修正。其他語意缺陷的總數尚未確定；機械檢查已發現的未修復錯誤為 0，不能推論語意問題為 0。
- 新增：本報告。修改：`docs/operations/configuration.md`。合併、遷移、刪除：無。
- 原始資訊對應：兩段原文仍在原檔，唯一變動為連結文字與空格；引用目標維持不變。

## 已執行的整改

- `docs/operations/configuration.md`：將兩段整段包裹的連結改為短文字引用，保留原有部署條件、Secret 版本說明與目標頁。
- 本輪未合併、遷移或刪除文件；沒有將無法核實的規則搬移到其他權威頁。

## 20 項驗收結果

`P` = PASS，`B` = BLOCKED，`N` = N/A。矩陣的 20 個字元由左至右對應下表 01–20。`B` 的共同阻礙是本次單一聊天無法完整人工核對 120,940 行及其對應實作；已嘗試 Repomix 全量打包、逐檔機械掃描、既有檢查及圖表渲染。解除條件是持續逐頁建立規則與實作對照，且在穩定工作樹上重跑全量驗證；DDD 發現需改寫時另受唯讀範圍約束。

| #   | 標準            | 目前結果與依據                                                                        |
| --- | --------------- | ------------------------------------------------------------------------------------- |
| 01  | 唯一權威        | B：未逐條完成跨頁規則語意比對，不能推定唯一權威。                                     |
| 02  | 分類            | P：依現有目錄分類逐檔建索引；上游快照歸參考。                                         |
| 03  | 命名            | P（自有）：既有命名檢查通過；上游檔保持來源名稱。                                     |
| 04  | 目錄            | B：已盤點位置，但未逐頁驗證領域歸屬與最佳放置。                                       |
| 05  | 領域語言        | B：未逐條比對 DDD 領域詞彙及跨頁定義。                                                |
| 06  | 標題            | P（自有）：逐檔檢查 H1 與標題跳級；上游格式不重寫。                                   |
| 07  | Markdown        | P（自有）：既有 Prettier／markdownlint 通過。                                         |
| 08  | Prettier        | P（自有）：Prettier check 通過；DDD 因唯讀限制受阻；上游依規範排除。                  |
| 09  | markdownlint    | P（自有）：markdownlint 0 issue；DDD 因唯讀限制受阻；上游依規範排除。                 |
| 10  | 連結            | P（自有與 DDD）：既有 201 份文件的相對連結及錨點檢查通過；上游頁面排除。              |
| 11  | 程式碼區塊      | P（自有）：逐檔掃描無未標語言的 fenced block。                                        |
| 12  | Mermaid         | B（含 Mermaid 頁）：90 段語法通過，但流程語意未逐圖與實作／決策比對；無圖頁 N。       |
| 13  | Front Matter    | N（自有）：自有文件未使用 front matter，狀態由分類或正文表達；上游保留來源 metadata。 |
| 14  | OpenAPI／Schema | B：未逐頁核對 API／Schema 敘述是否完全引用現有契約。                                  |
| 15  | 文件與實作      | B：未逐條完成文件、原始碼與現行部署配置比對。                                         |
| 16  | CI              | P（自有）：使用既有文件、設定檢查；不推定未涵蓋的業務驗收。                           |
| 17  | Git／ADR        | B：ADR／Git 歷史存在，但未逐頁驗證決策脈絡與現況區分。                                |
| 18  | 生命週期        | B：提案／ADR 有分類，其他頁的 Draft／Active／Deprecated 語意尚未逐頁核對。            |
| 19  | AGENTS 範圍     | P：已讀根部及 docs/AGENTS；DDD 路徑列入唯讀保護。                                     |
| 20  | 審查／所有權    | B：文件規範按模組要求審查，但未逐頁核實維護責任與實際審查證據。                       |

## 逐檔矩陣

`P/B/N` 僅代表上表所述的已驗證範圍。上游快照不作本專案權威規則，除分類外標 `N`；DDD 為唯讀且既有檢查排除該目錄，排版、lint 與完整 CI 驗收在本任務標 `B`。更新：已對全部 69 份 DDD 文件直接執行 Prettier 與 markdownlint。16 份排版有差異；design-diagrams.md 第 17 行從 H1 跳至 H3（MD001）。這些修正受唯讀要求限制。逐檔第 08、09 欄以本次直接檢查為準。上游 manifest 的 176 筆 SHA-256 與本機文件逐一相符。

| 檔案                                                                                                      | 分類              | 01–20                  |
| --------------------------------------------------------------------------------------------------------- | ----------------- | ---------------------- |
| `docs/AGENTS.md`                                                                                          | AGENTS.md         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/architecture/data.md`                                                                               | architecture      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/architecture/runtime/backend-boundaries.md`                                                         | architecture      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/architecture/runtime/code-layout.md`                                                                | architecture      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/architecture/runtime/dependencies.md`                                                               | architecture      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/architecture/runtime/overview.md`                                                                   | architecture      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/architecture/runtime/react-lifecycle.md`                                                            | architecture      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/architecture/runtime/shared-adapters.md`                                                            | architecture      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/contracts/business-operations.md`                                                                   | contracts         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/contracts/http.md`                                                                                  | contracts         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/contracts/identity.md`                                                                              | contracts         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/contracts/organization-team-ownership.md`                                                           | contracts         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/decisions/0002-baseline.md`                                                                         | decisions         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/decisions/0003-cloud-run-primary.md`                                                                | decisions         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/decisions/0004-monorepo.md`                                                                         | decisions         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/environment/codex-cloud.md`                                                             | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/environment/troubleshooting.md`                                                         | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/environment/verification.md`                                                            | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/agents/codex.md`                                                                | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/agents/context7.md`                                                             | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/agents/google.md`                                                               | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/agents/skills.md`                                                               | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/api/change-api.md`                                                              | development       | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/development/tooling/backend/firestore.md`                                                           | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/backend/go-development.md`                                                      | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/backend/gopls.md`                                                               | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/frontend/nextjs-devtools.md`                                                    | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/frontend/rich-menu-layout.md`                                                   | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/frontend/shadcn-reference/LICENSE.md`                                           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/INDEX.md`                                    | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/manifest.json`                               | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/chain-of-thought.md`  | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/chat.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/checkpoint.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/confirmation.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/context.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/file-tree.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/inline-citation.md`   | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/message.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/model-selector.md`    | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/plan.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/prompt-input.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/queue.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/reasoning.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/sources.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/stack-trace.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/subagent.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/suggestion.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/task.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/terminal.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/test-results.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/base/tool.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/chain-of-thought.md` | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/chat.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/checkpoint.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/confirmation.md`     | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/context.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/file-tree.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/inline-citation.md`  | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/message.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/model-selector.md`   | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/plan.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/prompt-input.md`     | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/queue.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/reasoning.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/sources.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/stack-trace.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/subagent.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/suggestion.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/task.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/terminal.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/test-results.md`     | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements/radix/tool.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/ai-elements.md`                        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/accordion.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/alert-dialog.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/alert.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/aspect-ratio.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/attachment.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/avatar.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/badge.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/breadcrumb.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/bubble.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/button-group.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/button.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/calendar.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/card.md`               | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/carousel.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/chart.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/checkbox.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/code-block.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/collapsible.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/combobox.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/command.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/context-menu.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/data-table.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/date-picker.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/dialog.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/direction.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/drawer.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/dropdown-menu.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/empty.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/field.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/hover-card.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/input-group.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/input-otp.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/input.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/item.md`               | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/kbd.md`                | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/label.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/marker.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/menubar.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/message-scroller.md`   | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/message.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/native-select.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/navigation-menu.md`    | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/pagination.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/popover.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/progress.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/questionnaire.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/radio-group.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/resizable.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/scroll-area.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/select.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/separator.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/sheet.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/sidebar.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/skeleton.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/slider.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/spinner.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/switch.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/table.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/tabs.md`               | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/textarea.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/toast.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/toggle-group.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/toggle.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/tooltip.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/trace.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/base/typography.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/accordion.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/alert-dialog.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/alert.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/aspect-ratio.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/attachment.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/avatar.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/badge.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/breadcrumb.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/bubble.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/button-group.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/button.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/calendar.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/card.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/carousel.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/chart.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/checkbox.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/code-block.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/collapsible.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/combobox.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/command.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/context-menu.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/data-table.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/date-picker.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/dialog.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/direction.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/drawer.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/dropdown-menu.md`     | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/empty.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/field.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/hover-card.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/input-group.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/input-otp.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/input.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/item.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/kbd.md`               | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/label.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/marker.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/menubar.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/message-scroller.md`  | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/message.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/native-select.md`     | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/navigation-menu.md`   | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/pagination.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/popover.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/progress.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/questionnaire.md`     | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/radio-group.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/resizable.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/scroll-area.md`       | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/select.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/separator.md`         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/sheet.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/sidebar.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/skeleton.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/slider.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/spinner.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/switch.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/table.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/tabs.md`              | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/textarea.md`          | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/toast.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/toggle-group.md`      | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/toggle.md`            | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/tooltip.md`           | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/trace.md`             | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components/radix/typography.md`        | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn-reference/snapshot/pages/components.md`                         | upstream-snapshot | `NPNNNNNNNNNNNNNNNNNN` |
| `docs/development/tooling/frontend/shadcn.md`                                                             | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/shared/gcloud.md`                                                               | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/shared/github.md`                                                               | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/shared/graphify.md`                                                             | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/shared/mermaid.md`                                                              | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/shared/mise.md`                                                                 | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/shared/repomix.md`                                                              | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/tooling/shared/rich-menu-generator.md`                                                  | development       | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/development/tooling/testing/commands.md`                                                            | development       | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/development/tooling/testing/firestore-emulator.md`                                                  | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/verification/evidence.md`                                                               | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/verification/document-audit-2026-11-24.md`                                              | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/verification/standardization-backlog.md`                                                | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/verification/standardization.md`                                                        | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/capabilities/architect.md`                                                     | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/capabilities/documentation.md`                                                 | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/capabilities/explorer.md`                                                      | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/capabilities/implementer.md`                                                   | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/capabilities/planning.md`                                                      | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/capabilities/reviewer.md`                                                      | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/capabilities/tester.md`                                                        | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/handoff.md`                                                                    | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/implementation.md`                                                             | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/development/workflow/publication.md`                                                                | development       | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/domains/agent/design.md`                                                                            | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/attendance/design.md`                                                                       | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/discussion/design.md`                                                                       | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/enterprise/design.md`                                                                       | domains           | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/domains/issue/design.md`                                                                            | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/journal/journal.md`                                                                         | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/knowledge/design.md`                                                                        | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/membership/design.md`                                                                       | domains           | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/domains/object/design.md`                                                                           | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/operations/design.md`                                                                       | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/organization/design.md`                                                                     | domains           | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/domains/project/project.md`                                                                         | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/repository/design.md`                                                                       | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/summary/summary.md`                                                                         | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/team/design.md`                                                                             | domains           | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/domains/weather/weather.md`                                                                         | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/domains/workflow/design.md`                                                                         | domains           | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/integrations/firestore/browser-session.md`                                                          | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/firestore/connection.md`                                                               | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/firestore/consistency.md`                                                              | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/firestore/data-model.md`                                                               | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/firestore/lifecycle.md`                                                                | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/firestore/querying.md`                                                                 | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/geolocation/browser-api.md`                                                            | integrations      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/integrations/google-cloud/background-jobs.md`                                                       | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/google-cloud/build-and-release.md`                                                     | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/google-cloud/document-processing.md`                                                   | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/google-cloud/observability.md`                                                         | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/google-cloud/recaptcha.md`                                                             | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/google-cloud/reporting.md`                                                             | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/google-cloud/workload-identity.md`                                                     | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/capabilities.md`                                                                  | integrations      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/integrations/line/commands.md`                                                                      | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/configuration.md`                                                                 | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/group-compatibility.md`                                                           | integrations      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/integrations/line/journey.md`                                                                       | integrations      | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/integrations/line/mini-app-identity.md`                                                             | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/presentation.md`                                                                  | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/setup.md`                                                                         | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/sources.md`                                                                       | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/webhook.md`                                                                       | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/line/wiring.md`                                                                        | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/integrations/storage/cloud-storage.md`                                                              | integrations      | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/operations/cloud-run.md`                                                                            | operations        | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/operations/configuration.md`                                                                        | operations        | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/operations/container-build.md`                                                                      | operations        | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/operations/release-acceptance.md`                                                                   | operations        | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/operations/rich-menu.md`                                                                            | operations        | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/operations/troubleshooting.md`                                                                      | operations        | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/proposals/domain-collaboration.md`                                                                  | proposals         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/proposals/firestore-data-contracts.md`                                                              | proposals         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/proposals/line-engineering/design.md`                                                               | proposals         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/proposals/membership.md`                                                                            | proposals         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/proposals/platform/delivery.md`                                                                     | proposals         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/proposals/platform/domain-model.md`                                                                 | proposals         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/proposals/platform/package-layout.md`                                                               | proposals         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/proposals/webhook-delivery/design.md`                                                               | proposals         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/references/ddd/README.md`                                                                           | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/adapter.md`                                                                          | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/aggregate-root.md`                                                                   | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/aggregate.md`                                                                        | ddd-protected     | `BPPBBBBBPPBBNBBBBBPB` |
| `docs/references/ddd/anti-corruption-layer.md`                                                            | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/application-service.md`                                                              | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/architecture-testing.md`                                                             | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/bounded-context-integration.md`                                                      | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/bounded-context.md`                                                                  | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/command.md`                                                                          | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/composition-root.md`                                                                 | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/consistency.md`                                                                      | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/context-map.md`                                                                      | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/context-mapping-patterns.md`                                                         | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/context-mapping.md`                                                                  | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/core-subdomain.md`                                                                   | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/cqrs.md`                                                                             | ddd-protected     | `BPPBBBBBPPBBNBBBBBPB` |
| `docs/references/ddd/dependency-inversion.md`                                                             | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/dependency-rule.md`                                                                  | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/design-diagrams.md`                                                                  | ddd-protected     | `BPPBBBBPBPBBNBBBBBPB` |
| `docs/references/ddd/design-guide.md`                                                                     | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/design-process.md`                                                                   | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/domain-authorization.md`                                                             | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/domain-command.md`                                                                   | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-contract.md`                                                                  | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/domain-error.md`                                                                     | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-event-handler.md`                                                             | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-event.md`                                                                     | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/domain-model-review.md`                                                              | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/domain-model.md`                                                                     | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-modeling.md`                                                                  | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-policy.md`                                                                    | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-primitive.md`                                                                 | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-service.md`                                                                   | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain-testing.md`                                                                   | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/domain.md`                                                                           | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/driven-adapter.md`                                                                   | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/driven-port.md`                                                                      | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/driving-adapter.md`                                                                  | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/driving-port.md`                                                                     | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/entity.md`                                                                           | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/event-sourcing.md`                                                                   | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/eventstorming.md`                                                                    | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/eventual-consistency.md`                                                             | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/factory.md`                                                                          | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/generic-subdomain.md`                                                                | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/hexagonal-architecture.md`                                                           | ddd-protected     | `BPPBBBBBPPBBNBBBBBPB` |
| `docs/references/ddd/idempotency.md`                                                                      | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/identity.md`                                                                         | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/integration-event.md`                                                                | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/invariant.md`                                                                        | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/model-evolution.md`                                                                  | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/optimistic-concurrency.md`                                                           | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/outbox-pattern.md`                                                                   | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/port.md`                                                                             | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/process-manager.md`                                                                  | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/query.md`                                                                            | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/repository.md`                                                                       | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/saga.md`                                                                             | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/specification.md`                                                                    | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/state-machine.md`                                                                    | ddd-protected     | `BPPBBBBBPPBNNBBBBBPB` |
| `docs/references/ddd/subdomain.md`                                                                        | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/supporting-subdomain.md`                                                             | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/transaction.md`                                                                      | ddd-protected     | `BPPBBBBPPPBBNBBBBBPB` |
| `docs/references/ddd/ubiquitous-language.md`                                                              | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/unit-of-work.md`                                                                     | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/upstream-downstream.md`                                                              | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/use-case.md`                                                                         | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/references/ddd/value-object.md`                                                                     | ddd-protected     | `BPPBBBBPPPBNNBBBBBPB` |
| `docs/standards/design.md`                                                                                | standards         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/standards/documentation.md`                                                                         | standards         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/standards/engineering.md`                                                                           | standards         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/standards/go.md`                                                                                    | standards         | `BPPBBPPPPPPBNBBPBBPB` |
| `docs/standards/miniapp-ui.md`                                                                            | standards         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/standards/module-constraints.md`                                                                    | standards         | `BPPBBPPPPPPNNBBPBBPB` |
| `docs/standards/naming.md`                                                                                | standards         | `BPPBBPPPPPPNNBBPBBPB` |

## 未解除事項與續接條件

1. 對 115 份自有文件逐條整理業務規則與權威來源，與 69 份唯讀 DDD、OpenAPI、Schema 和現行程式比對；修正可證實且位於可寫範圍的矛盾。
2. 對 90 段 Mermaid 分別核對流程語意；對上游快照保留原格式並明確區分參考與專案權威。
3. 每輪實際修正後重新執行 Repomix、文件檢查、連結檢查與範圍差異檢查。
4. DDD 若發現需要修改，因本任務唯讀限制只能記錄，不得直接整改；解除需另有明確變更授權。

本次尚未滿足「所有可修復問題為零」及「完整語意驗收」；最終狀態為**部分完成／受阻**。
