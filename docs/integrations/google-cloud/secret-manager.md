# Secret Manager 簡要規格

**狀態：部署文件已有規範；雲端狀態另驗收。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

保存 LINE Token、Channel Secret 及服務端 API 機密，支援版本與輪替。

建立機密版本 → 以 Cloud Run reference 提供執行環境 → 驗證新 revision → 按計畫輪替與停用舊版本。

## 資料、授權與失敗邊界

按單一 secret 授予必要存取；原始值不進 Git、映像、建置替代參數、回應或日誌。

## 導入條件

機密名稱、版本來源、讀取身分、輪替負責人及回復界線由部署設定維護；瀏覽器公開金鑰另依 API 限制管理。

## 驗收

無權讀取、錯誤／停用版本、輪替及回復均有明確結果，日誌不含 payload。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/secret-manager/docs/overview)，查證範圍見[整合總覽](README.md#sources)。[部署規格](../../operations/cloud-run.md) · [IAM](iam.md)。

[Google 服務索引](README.md)
