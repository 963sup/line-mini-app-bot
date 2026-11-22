# IAM 與服務帳戶 簡要規格

**狀態：部署文件已有規範；雲端狀態另驗收。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

控制建置、部署、應用執行及工作觸發對 Google 資源的存取。

列出身分 → 必要行為 → 目標資源 → 最小權限；按環境建立與驗證授權。

## 資料、授權與失敗邊界

分開建置、執行與工作呼叫身分，優先使用附掛身分及短效憑證；IAM 不替代 LINE 使用者的領域授權。

## 導入條件

避免以 Owner／Editor 或長期服務帳戶金鑰作為預設；記錄 impersonation、資源範圍及撤權方式。

## 驗收

缺權被拒、跨環境／跨資源隔離、撤權後行為與 worker 呼叫身分均符合預期。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/iam/docs/overview)，查證範圍見[整合總覽](README.md#sources)。[領域授權](../../references/ddd/domain-authorization.md) · [部署規格](../../operations/cloud-run.md)。

[Google 服務索引](README.md)
