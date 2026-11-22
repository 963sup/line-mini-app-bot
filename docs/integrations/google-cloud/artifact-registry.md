# Artifact Registry 簡要規格

**狀態：已有部署流程來源；目標環境另驗收。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

保存建置映像，讓部署與回復能定位確切產物。

Build 推送映像 → 記錄 digest 與來源 → Cloud Run 使用已驗證映像 → 依保存政策清理。

## 資料、授權與失敗邊界

Writer、Reader 與刪除權限分開；清理前核對目前服務及保留回復版本仍需的映像。

## 導入條件

以 digest 追蹤實際部署，tag 僅作標記；掃描結果、保留期限與例外按發布需求管理。

## 驗收

未授權推送／讀取被拒，digest 可追到 build，清理不破壞已承諾的回復路徑。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/artifact-registry/docs/repositories)，查證範圍見[整合總覽](README.md#sources)。[Cloud Build](cloud-build.md) · [發布驗收](../../operations/release-acceptance.md)。

[Google 服務索引](README.md)
