# 0003：Cloud Run 是正式應用執行環境

- 狀態：Accepted
- 日期：2026-11-08

## 背景

應用依賴 Google Cloud 資源。Cloud Run 可透過執行服務帳戶與 Application Default Credentials 存取資源，降低長期憑證管理。

## 決策

- 正式應用執行於 Cloud Run。
- 使用單一容器映像與明確 runtime 設定契約。
- 使用專用服務帳戶及最小 IAM。
- 機密優先以 Secret Manager reference 注入，不寫入映像或儲存庫。

## 後果

部署與回退以 Cloud Run revision 為單位。發布證據必須包含 project、region、service、revision、執行帳戶、機密 reference、流量及核心流程；本 ADR 本身不授權建立或修改雲端資源。
