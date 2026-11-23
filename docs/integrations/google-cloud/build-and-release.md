# 建置與映像發布

<a id="cloud-build"></a>

## Cloud Build

**狀態：已有部署流程來源；目標環境另驗收。**

**用途與最小流程**

執行受控建置、檢查、映像推送及 Cloud Run 發布。

指定來源 revision → 執行必要檢查 → 建置映像 → 推送 Registry → 依發布門檻部署。

**資料、授權與失敗邊界**

建置身分與應用執行身分分離；只引用必要機密，保留來源、build ID 與映像 digest。

**導入條件**

現行 trigger 是否使用版本庫 build 檔須另查；建置成功不代表功能與 LINE 整合驗收。

**驗收**

錯誤來源、檢查失敗、推送失敗與部署失敗不被視為成功，產物可追溯來源。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/build/docs/overview)。[部署規格](../../operations/cloud-run.md) · [Artifact Registry](#artifact-registry)。

<a id="artifact-registry"></a>

## Artifact Registry

**狀態：已有部署流程來源；目標環境另驗收。**

**用途與最小流程**

保存建置映像，讓部署與回復能定位確切產物。

Build 推送映像 → 記錄 digest 與來源 → Cloud Run 使用已驗證映像 → 依保存政策清理。

**資料、授權與失敗邊界**

Writer、Reader 與刪除權限分開；清理前核對目前服務及保留回復版本仍需的映像。

**導入條件**

以 digest 追蹤實際部署，tag 僅作標記；掃描結果、保留期限與例外按發布需求管理。

**驗收**

未授權推送／讀取被拒，digest 可追到 build，清理不破壞已承諾的回復路徑。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/artifact-registry/docs/repositories)。[Cloud Build](#cloud-build) · [發布驗收](../../operations/release-acceptance.md)。

[Google 服務索引與查證範圍](./)
