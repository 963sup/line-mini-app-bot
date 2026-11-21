# 版本控制與 PR 交付

需發布已驗證的變更時，按本頁核對分支、PR、已驗證的提交與發布授權。

## Repository publication

**開發與維運指引**

```mermaid
flowchart TB
  A["從最新 main 開始，核對遠端、工作樹與差異"] --> B["只提交本任務範圍並完成相關驗證"]
  B --> C{"任務範圍允許發布？"}
  C -->|否| L["依指定範圍交付"]
  C -->|是| P["推送 work，建立或更新指向 main 的 PR"]
  P --> V{"PR head 等於已驗證 SHA，且通過審查與分支保護？"}
  V -->|否| W["處理檢查或審查結果後重新確認"]
  V -->|是| M["Squash and merge"]
  M --> S["fetch 最新兩分支，將 origin/main 合回 work 並正常推送"]
  S --> E["核對祖先關係與差異，交付 PR、main 提交及驗證結果"]
```

`work` 可保存多筆開發提交，`main` 每項可獨立回退的變更保留一筆提交，可涵蓋同一功能的前端、API 與後端。PR 描述最終行為、驗證和限制。同步時若他人已推進分支，先核對新增內容。版本驗收後可按發布安排加標籤。重寫已發布歷史需單次明確授權與可還原備份，遠端拒絕時遵循分支保護與審核結果。

GitHub MCP 與 Git CLI 的使用見 [GitHub 工具指南](../tooling/mcp/github.md)。推送、CI 與正式部署分別核對，見 [發布驗收](../../guides/release/release-acceptance.md)。

[開發流程索引](README.md)
