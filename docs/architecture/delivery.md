# 實作階段與驗收

> 狀態：規劃；不預建全部套件、不宣稱現有 health server 已實作以下功能。

| 階段 | 範圍 | 驗收 |
| --- | --- | --- |
| 1 | 一對一、Webhook、Inbox、Flex Reply | 原始 body 驗簽、多事件、空事件、重送、Unsend、無效輸入 |
| 2 | 群組事件、群組綁定、角色授權 | 缺少 actor、越權、退出重入、私人內容不回群 |
| 3 | Go MINI App 頁面、LIFF、session、API | 錯誤 channel、過期 token、跨使用者／群組、CSRF |
| 4 | Outbox、Worker、Garr | 重啟、租約過期、多副本競爭、429、超時結果不明、上限與背壓 |
| 5 | Rich Menu、Flex 管理 | payload／altText、不可變版本、測試發送、發布回復、裝置顯示 |
| 6 | 投產 | 持續 Worker 資源、IAM、索引、額度、保存／刪除政策、告警與復原 |

核心用例以單元測試驗證成功與拒絕；HTTP adapter 使用受控 HTTP 測試；Firestore 使用 opt-in emulator 整合測試。
真實渠道驗證群組加入設定、Push 資格、Rich Menu、LIFF；emulator 不代表 production IAM／索引可用。
Service Message 等正式資格功能需另外驗收後啟用。

## Codex 任務輸入與完成條件

每次只讀 [overview](overview.md) 路由所需文件，再讀當前範圍的原始碼。
交接列出目標、擁有檔案、依賴與驗收；一檔一寫入者。小任務不要自動展開全部角色。
新增 package 前依 [layout](layout.md) 選位置，依 [dependencies](dependencies.md) 核對 import 方向。
完成時列出實際變更、執行過的檢查、未實作部分與外部阻礙；Graphify/Repomix 的產出不是功能測試。
