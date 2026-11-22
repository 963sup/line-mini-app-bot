# Cloud Run functions 簡要規格

**狀態：候選方案；不是現行主服務部署替代品。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

處理獨立的檔案事件、縮圖、metadata 擷取或輕量整合入口；若既有 Go 服務可簡單處理則沿用。

HTTP 或 CloudEvent 觸發 → 驗證來源及操作身分 → 冪等處理指定資源版本 → 記錄結果。

## 資料、授權與失敗邊界

明確選定執行環境與觸發身分，不把 Firebase Authentication／使用者權限當作自動具備；業務異動仍經相同規則或授權用例。

## 導入條件

依[Cloud Run 決策](../../decisions/0003-cloud-run-primary.md)評估附加 function 的部署與所有權；不是自動拆服務。事件可能重送，限制重試與寫回循環，較長批次改評估 Jobs。

## 驗收

錯誤身分、事件重送、部分失敗、逾時與舊檔案版本不造成重複異動，必要業務操作仍可驗收。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/run/docs/functions/overview)，查證範圍見[整合總覽](README.md#sources)。[Eventarc](eventarc.md) · [Jobs](cloud-run-jobs.md)。

[Google 服務索引](README.md)
