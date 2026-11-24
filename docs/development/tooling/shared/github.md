# GitHub 共用操作

查閱儲存庫、分支、提交、Issue、PR 與 CI 時，優先使用當前聊天曝光的 GitHub MCP。專案服務透過 `GITHUB_MCP_TOKEN` 驗證，實際能力須依主機工具 schema 與目標儲存庫權限核對。提交、推送與授權依[發布流程](../../workflow/publication.md#repository-publication)。

需保留已驗證的本地 SHA 時，先確認遠端已有該提交，再使用可指定 SHA 的 ref 更新工具。重新寫檔或建立提交可能產生不同 SHA。使用快轉更新；競爭更新保護如 `expected_sha` 須核對實際工具 schema。[GitHub REST ref 更新](https://docs.github.com/en/rest/git/refs#update-a-reference) · [MCP 診斷](../../environment/troubleshooting.md#設定與-mcp-診斷)

## 確認來源與交付狀態

1. 確認目標 owner/repository、分支與提交 SHA，區分本地工作樹、遠端分支與 PR 的來源。
2. 查閱相關差異、Issue、PR 與 CI；CI 結果必須對應實際提交，不以其他提交的成功代替。
3. 依任務授權執行必要操作，發布順序及工具選擇遵循發布流程。工作區有未提交變更時保留它們。
4. 寫入後讀回遠端分支、提交或頁面；分別回報提交、推送、CI 與部署狀態。

Git CLI 的本地狀態與遠端資料不能互相替代；工具回傳成功也不表示應用已驗收。

## GitHub Actions／PR 檢查失敗

使用者要求「修理紅燈」、「CI 失敗」或「PR 檢查失敗」時，MUST 優先使用當前聊天可用的 GitHub MCP 查證，再依[除錯流程](../../workflow/implementation.md#diagnostic-loop)修正。

1. 確認儲存庫、PR／分支、失敗提交 SHA、workflow run／attempt、job 與 step。讀取對應檢查結果及失敗日誌，不以舊提交的失敗推斷目前狀態。
2. 依日誌區分程式回歸、測試、建置、設定、權限或外部服務問題；對照該提交的來源與工作區差異，定位根因後做最小必要修正。不得以停用檢查、削弱測試或反覆重跑掩蓋失敗。
3. 執行對應本機驗證；提交、推送及重跑依本次授權與[發布流程](../../workflow/publication.md#repository-publication)。重跑前確認原因與目標，不盲目重跑全部 workflow。
4. 透過 GitHub MCP 讀回修復後提交的檢查狀態，核對 SHA、run／attempt 與必要檢查。只有目標提交的必要檢查實際成功，才回報紅燈解除；pending、cancelled、skipped 或只有本機通過都不能替代成功證據。
5. 交付列出失敗原因、修改、實際測試、修復提交 SHA、CI 連結及剩餘阻塞。尚未推送或遠端仍執行中時，明確標示未完成遠端驗收。

MCP 未曝光或缺少 Actions 日誌／重跑能力時，先指出具體限制，再使用任務授權範圍內可用的 GitHub CLI／API。工具名稱與能力以當前 schema 為準，不假設每個 MCP 都提供同一組功能。

## Wiki publication

Inspect existing pages and preserve their language, navigation and unrelated content. Use a Wiki-specific tool when available, otherwise use the repository’s separate `.wiki.git` remote. Review the diff, publish only the requested changes without force, and verify the published page or remote commit.

Wiki 發布需要明確授權，不能由一般文件修改或儲存庫權限推定。

[工具索引](../)
