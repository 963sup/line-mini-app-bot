# ADR 0003：Cloud Run 是唯一應用部署目標

狀態：Accepted

日期：2026-11-08

本決策取代先前的部署平台決策；其餘應用架構與
「先基礎、再接入、最後出勤」的交付順序維持不變。

## 背景

應用使用 Firestore。Cloud Run 可透過執行服務帳戶與 Application Default
Credentials 存取 Google Cloud 資源，能減少跨平台長期憑證的管理工作。

## 決策

- Cloud Run 是唯一正式應用執行環境。
- 應用使用單一容器映像與環境變數契約。
- Cloud Run 使用專用服務帳戶與最小 Firestore IAM；LINE 機密由部署環境安全注入，
  優先使用 Secret Manager 綁定，不寫入映像或版本庫。

## 現況與完成條件

版本庫已有 Dockerfile；Cloud Run 的服務、服務帳戶、IAM、
機密綁定及自動部署 workflow 尚未建立或驗收。

主要部署完成需記錄 Cloud Run service URL、region、revision、服務帳戶、必要 IAM、
機密來源及 LINE Webhook 驗證結果。本 ADR 不授權建立雲端資源或執行部署。
