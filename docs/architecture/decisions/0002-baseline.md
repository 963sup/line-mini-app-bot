# ADR 0002：基礎出勤與分階段落地基線

狀態：Accepted

本 ADR 的舊部署平台部分已由 [ADR 0003](0003-cloud-run-primary.md) 取代；
其餘應用基線仍有效。本 ADR 取代 ADR 0001「持續 Worker 必須隨首版部署」的假設。
平台擴充契約保留為後續參考，遇到範圍或交付順序衝突時以本紀錄為準。

## 背景

現有架構已明確定義程式分層；部署決策後續更新為只使用 Cloud Run。
先穩定基礎及暫不展開測試工作的順序不變。

## 決策

- Go 模組化單體，一個容器部署至 Cloud Run。
- 採輕量 Hexagonal：HTTP 入口、業務使用案例與外部 adapter 分工，在 main 組裝依賴。
- Webhook 與 MINI App 共用業務使用案例；業務層不依賴 LINE／Firestore SDK 型別。
- Firestore 為持久化方向；後端負責授權，外部發送不放入 transaction callback。
- Cloud Run 使用執行服務帳戶與最小 IAM，機密由 Secret Manager 綁定。
- 不預建未確認業務模組，不引入微服務、完整 CQRS、Event Sourcing 或通用 Repository 框架。

## 階段範圍

首版僅做 [基礎出勤](../../features/attendance.md)：上班、下班與本人紀錄。
首期入口、人員資格、時間歸屬與重複操作規則集中在功能規格確認；補登不納入首版。
依 [分階段落地](../../development/foundation.md) 先完成執行與部署基礎、再完成資料與 LINE 接入，
最後實作出勤；不一步到位。不預設會員管理、排班、薪資或通知功能。

目前先穩定工具設定、應用骨架、設定注入與部署說明；不新增測試方案、案例或工具。
此決策不要求移除現有測試或 CI，也不表示既有檢查已執行或通過。

## 後果與變更條件

API、資料欄位與功能細節在此基線內逐步補齊，不因尚未展開細節而重開架構。
只有明確的新需求或實測證據需要改變基線時，才新增 ADR，說明原因與影響。
本紀錄不表示 LINE／Firestore 已完成整合或正式部署。

## 查核

文件中的部署目標應一致指向 Cloud Run。
參考：[應用架構](../overview.md)、[開發原則](../principles.md)、
[Cloud Run](../../operations/cloud-run.md)。
