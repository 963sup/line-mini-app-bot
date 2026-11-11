# 後端功能邊界

第三階段已完成 HTTP、領域與持久化格式的分離；公開 OpenAPI、資料文件與交易語意保持相容。

| 目錄 | 責任與允許依賴 |
| --- | --- |
| business/attendance/{punch,records,worktime,corrections,approvals,leave,board} | 各子功能的模型、規則與能力契約；依賴方向見 [重構對照](../development/business-refactor.md) |
| business/attendance/http | 出勤 HTTP、生成 DTO 轉換、認證後的契約驗證；依賴消費端 Authenticator 介面 |
| business/attendance/firestore | 文件 DTO、查詢／交易、錯誤轉譯、固定版本防重編碼 |
| business/attendance/line、business/attendance/presentation | 聊天流程、出勤卡片與不依賴 SDK 的共用呈現 |
| business/summary/messagecollection、business/summary/messagecollection/line | 訊息模型與收集契約；LINE 文字收集、編輯與收回，不依賴摘要生成 |
| business/summary/line | 收集與摘要指令的組合入口，先處理指令，再靜默收集一般文字 |
| business/summary/summarization | 摘要用例與能力契約；不依賴 SDK |
| business/summary/summarization/line、business/summary/summarization/cloudtasks | 摘要指令、佇列排程與結果傳送 |
| business/summary/summarization/http | 摘要工作 HTTP，透過 WorkerVerifier 驗證工作身分 |
| business/summary/firestore | 收集與摘要共用的文件、版本狀態與交易 |
| business/weather/forecast/cwa、business/weather/forecast/line | 預報來源與聊天呈現；組裝由 server 負責 |
| identity | 身分資料、Verifier 契約及 request context；不依賴 HTTP／SDK |
| identity/http | Bearer 輸入、身分回應、HTTP 錯誤與 MINI App 設定入口 |
| identity/line、identity/google | LINE 本人身分與 Google 工作身分驗證的外部整合 |
| browser、diagnostics/geolocation | 靜態頁面／CSP，以及獨立定位診斷 |
| httpapi | 生成介面接線與 spec，不 import 功能實作 |
| transport/http、transport/line | 共用 HTTP 格式與 LINE 傳輸；不依賴功能 |
| server | 設定、外部 client 建立、手動注入、路由與生命週期 |

HTTP adapter 不建立 Firestore／LINE／Google client，也不讀 gRPC 狀態碼。出勤、摘要的資料層在 output port 邊界將技術錯誤轉成所屬 ErrUnavailable 或既有業務錯誤；context cancellation 保留 Go 標準錯誤語意。技術原因保存在 adapter 的私有錯誤值，Unwrap 只公開領域類別，不讓上層依賴 SDK 錯誤。

出勤 Firestore 公開方法在回傳邊界轉譯錯誤；需統一處理多個提早回傳時，以具名 error 回傳搭配 defer 完成，不保留僅轉呼叫的 `xxxOperation` 方法。私有共用交易只供 adapter 內部使用，保留技術錯誤供公開邊界轉譯。
`PunchSubmitter`、`PunchResolver` 與 `GroupHistoryReader` 分別表達定位提交、可能寫入拒絕收據的操作解析，以及本人歷史查詢；讀取分支不因缺少定位提交能力而切換。


身分驗證與契約驗證仍依既有順序。已驗證身分只在同次 request context 共用；它不是跨請求登入快取。摘要 HTTP 依賴 WorkerVerifier 最小介面，由 server 注入 Google verifier。

## 格式與資料相容

領域模型不帶 JSON／Firestore tags。HTTP 使用 OpenAPI 生成 DTO；Firestore 使用私有 document structs 與明確轉換。打卡 Request 屬儲存憑證，已移出領域模型。

原文件欄位名稱、omitempty、巢狀位置、事件回執及核准稽核格式保持不變。更正防重雜湊的 JSON 格式是已持久化的 v1 契約：由 Firestore codec 保留原欄位順序、伺服器欄位排除、UTC 與空休息區間正規化，不能改用 HTTP DTO 或直接序列化核心型別。

HTTP 舊月報以遷移前捕捉的 JSON fixture 驗證。Firestore 相容測試使用獨立凍結的舊模型比對新文件，並確認舊防重回執可重播；測試僅對本機模擬器執行，沒有正式資料寫入或回填。

## 持續檢查

既有 Go suite 包含 architecture/TestDependencyBoundaries，阻擋：核心引入外部實作、核心帶序列化 tags、HTTP 直接引入 SDK 或 outgoing adapter、共用傳輸反向依賴功能，以及 OpenAPI 接線依賴功能實作。此檢查驗證直接依賴與結構，不取代行為或整合驗證。

前端已完成 React hooks/state；剩餘真實 LINE／正式環境整合驗收；此階段不新增框架、部署或變更資料庫索引。
