# 版本控制與發布

## Repository publication

依[根指引](../../../AGENTS.md)在 `main` 開發並保留既有變更。每筆提交須完整、可獨立審查與回退，包含相關的程式、契約、設定與文件；提交內容須經[適用驗收](../tooling/testing/commands.md#變更驗收矩陣)，並審閱 staged diff。提交訊息說明具體變更與必要的原因、驗證或限制。

推送須有任務授權，核對遠端與待推送提交，以快轉更新 `origin/main`；遠端前進時整合並重新驗證，分支保護阻擋時依其規則處理並回報。推送後核對遠端 SHA、追蹤適用的 GitHub Actions，回報未完成的驗收。已發布歷史的重寫須取得單次明確授權並備妥可還原備份；版本驗收後可依發布安排加標籤。

GitHub MCP 與 Git CLI 的使用見 [GitHub 工具指南](../tooling/shared/github.md)。推送、CI 與正式部署分別核對，見 [發布驗收](../../operations/release-acceptance.md)。

[開發流程索引](./)
