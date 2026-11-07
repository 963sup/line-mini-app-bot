# 實作階段與驗收

> 狀態：規劃；不預建全部套件、不宣稱現有 health server 已實作以下功能。

## 當前交付順序

依 [ADR 0002](decisions/0002-baseline.md)、[分階段落地](../development/foundation.md)
與 [基礎出勤](../features/attendance.md)：執行／Fly 骨架 → 資料與首期 LINE 接入 → 上班、下班與本人紀錄。
目前不新增測試方案、案例或工具，既有 CI 保留。

## 後續平台擴充參考

以下舊平台階段僅供未來需求確認後參考，不是目前的實作順序或必做清單。

| 階段 | 範圍 | 驗收 |
| --- | --- | --- |
| 1 | 一對一、Webhook、Inbox、Flex Reply | 原始 body 驗簽、多事件、空事件、重送、Unsend、無效輸入 |
| 2 | 群組事件、群組綁定、角色授權 | 缺少 actor、越權、退出重入、私人內容不回群 |
| 3 | Go MINI App 頁面、LIFF、session、API | 錯誤 channel、過期 token、跨使用者／群組、CSRF |
| 4 | Outbox、Worker、Garr | 重啟、租約過期、多副本競爭、429、超時結果不明、上限與背壓 |
| 5 | Rich Menu、Flex 管理 | payload／altText、不可變版本、測試發送、發布回復、裝置顯示 |
| 6 | 投產 | 持續 Worker 資源、IAM、索引、額度、保存／刪除政策、告警與復原 |

Service Message 等正式資格功能需另外驗收後啟用。

## Codex 任務輸入與完成條件

每次只讀 [overview](overview.md) 路由所需文件，再讀當前範圍的原始碼。
交接列出目標、擁有檔案、依賴與驗收；一檔一寫入者。小任務不要自動展開全部角色。
新增 package 前依 [layout](layout.md) 選位置，依 [dependencies](dependencies.md) 核對 import 方向。
完成時列出實際變更、執行過的檢查、未實作部分與外部阻礙；Graphify/Repomix 的產出不是功能測試。
