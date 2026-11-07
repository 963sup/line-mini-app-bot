# 基礎設施與分階段落地

先建好可執行、可部署、可連線的基礎，再實作 [基礎出勤](../features/attendance.md)。
沿用 [架構基線](../architecture/decisions/0002-baseline.md)與
[部署決策](../architecture/decisions/0003-cloud-run-primary.md)，應用只部署至 Cloud Run。
本文件是工作順序與完成條件，不表示資源已建立，也不授權建立雲端資源或部署。
目前不新增測試方案、案例或工具，保留既有 CI。

## 目前狀態

| 項目 | 版本庫可確認的狀態 | 尚未證明的事項 |
| --- | --- | --- |
| 本機工具 | 有 mise 設定與 bootstrap／doctor 指令 | 乾淨環境是否可完整重建 |
| HTTP 骨架 | 有 cmd/server、可選 LINE 接入與正常關閉，沒有健康檢查路由 | 目標環境實際啟動狀態 |
| 容器與部署 | 有 Dockerfile；尚無 Cloud Run workflow | 容器實際建置、Cloud Run 資源、部署憑證及上線狀態 |
| 設定 | 已讀 PORT、Messaging API、MINI App 設定，不自動讀 .env | 真實渠道與環境對應尚未確認 |
| Firestore | 已有 cmd/firestorecheck 唯讀連線工具 | 資料庫、身分、IAM 與真實連線尚未確認；應用持久化未實作 |
| LINE | 已有 Webhook、1:1／群組、Flex／Postback、MINI App 驗證及 Rich Menu CLI | 真實收發、登入、選單設定與 MINI App 發布尚未確認 |
| 基礎出勤 | 已確認首版方向 | 功能程式尚未實作 |

不以設定檔存在、工具索引成功或依賴已下載，推論服務已可用。

## 第一階段：執行與部署基礎（目前優先）

1. 依 [環境設定](setup.md) 準備工具，整理啟動所需的設定與注入來源。
2. 啟動現有 HTTP 骨架，記錄 PORT、啟動指令、實際路由與關閉方式，不新增健康檢查路由。
3. 確認 Docker 可用，依現有 container-build 任務建置骨架；建置問題在此階段處理。
4. 依 [Cloud Run 指南](../operations/cloud-run.md) 選定 project、region、Artifact Registry、
   執行服務帳戶與機密來源。
5. 在取得部署授權及必要設定後，先部署 Cloud Run 骨架，記錄 URL、revision、
   服務帳戶、程序日誌與實際路由狀態。

完成條件：本機可啟動、容器可建置、Cloud Run 骨架可存取，並有對應執行紀錄。
尚未授權或缺憑證的項目標記阻塞，不把出勤功能混入這一階段。

## 第二階段：資料與 LINE 接入基礎

1. 決定 Firestore project、database 與位置，配置 Cloud Run 執行服務帳戶與最小 IAM。
2. 依 [Firestore 文件](../integrations/firestore.md) 配置應用所需的資料庫連線與認證。
3. 在應用中實作設定載入、缺值時的啟動錯誤、Firestore client 生命週期與基本連線。
4. 確認首期只採用的 LINE 入口及對應 Channel，依 [LINE 設定契約](../integrations/line-environment.md)
   注入必要設定，再接上身分驗證與必要的收發能力；不一次實作全部 LINE 能力。

完成條件：選定環境的資料連線、LINE 入口及可信人員識別各有實際操作紀錄；
失敗時能從日誌定位問題，且不記錄憑證與原始身分 token。
這是基礎連線確認，不新增自動化測試工程。

## 第三階段：基礎出勤

先收斂 [出勤規格](../features/attendance.md) 列出的必要決策，
再依序實作上班、下班與本人紀錄查詢。只新增這些操作需要的 API、資料欄位及 adapter。
完成條件：可識別本人並保存、查看正確的出勤紀錄，錯誤不會被顯示為成功。

基礎設施與接入尚未完成時，不宣稱出勤可以落地；文件中的完成條件也不是完成證據。
