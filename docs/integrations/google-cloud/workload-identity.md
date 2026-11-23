# 工作負載身分與機密

<a id="iam"></a>

## IAM 與服務帳戶

**狀態：部署文件已有規範；雲端狀態另驗收。**

**用途與最小流程**

控制建置、部署、應用執行及工作觸發對 Google 資源的存取。

列出身分 → 必要行為 → 目標資源 → 最小權限；按環境建立與驗證授權。

**資料、授權與失敗邊界**

分開建置、執行與工作呼叫身分，優先使用附掛身分及短效憑證；IAM 不替代 LINE 使用者的領域授權。

**導入條件**

避免以 Owner／Editor 或長期服務帳戶金鑰作為預設；記錄 impersonation、資源範圍及撤權方式。

**驗收**

缺權被拒、跨環境／跨資源隔離、撤權後行為與 worker 呼叫身分均符合預期。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/iam/docs/overview)。[領域授權](../../references/ddd/domain-authorization.md) · [部署規格](../../operations/cloud-run.md)。

<a id="secret-manager"></a>

## Secret Manager

**狀態：部署文件已有規範；雲端狀態另驗收。**

**用途與最小流程**

保存 LINE Token、Channel Secret 及服務端 API 機密，支援版本與輪替。

建立機密版本 → 以 Cloud Run reference 提供執行環境 → 驗證新 revision → 按計畫輪替與停用舊版本。

**資料、授權與失敗邊界**

按單一 secret 授予必要存取；原始值不進 Git、映像、建置替代參數、回應或日誌。

**導入條件**

機密名稱、版本來源、讀取身分、輪替負責人及回復界線由部署設定維護；瀏覽器公開金鑰另依 API 限制管理。

**驗收**

無權讀取、錯誤／停用版本、輪替及回復均有明確結果，日誌不含 payload。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/secret-manager/docs/overview)。[部署規格](../../operations/cloud-run.md) · [IAM](#iam)。

[Google 服務索引與查證範圍](README.md#sources)
