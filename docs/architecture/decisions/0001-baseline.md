# ADR 0001：應用架構基線與 Fly.io 部署

狀態：Accepted

## 背景

現有架構已明確定義程式分層；文件曾同時指向 Cloud Run 與 Fly.io，造成部署方向不一致。
使用者確認僅使用 Fly.io，並要求先穩定基礎，暫不展開測試工作。

## 決策

- Go 模組化單體，一個服務僅部署至 Fly.io。
- 採輕量 Hexagonal：HTTP 入口、業務使用案例與外部 adapter 分工，在 main 組裝依賴。
- Webhook 與 MINI App 共用業務使用案例；業務層不依賴 LINE／Firestore SDK 型別。
- Firestore 為持久化方向；後端負責授權，外部發送不放入 transaction callback。
- 正式環境機密使用 Fly Secrets。Fly 存取 Google 資源的身分與 IAM 必須另外配置。
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

文件中的部署目標與機密注入方式應一致指向 Fly.io／Fly Secrets。
參考：[應用架構](../overview.md)、[開發原則](../principles.md)、[Fly.io](../../operations/fly.md)。
