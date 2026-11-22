# api/contract/repository 目錄導覽

定義 Repository、空間、標籤、里程碑與統計等 API 契約。

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

從 [labels.yaml](labels.yaml)、[milestones.yaml](milestones.yaml)、[repositories.yaml](repositories.yaml)、[spaces.yaml](spaces.yaml)、[statistics.yaml](statistics.yaml) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：空間載入時序](../../../docs/domains/repository/README.md#空間載入時序)

Repository 列表與詳情的 `groupId` 只在群組擁有的儲存庫回傳。MINI App 使用此欄位建立出勤入口；出勤 API 仍依自己的規則驗證存取資格。

專案建立畫面以 `GET /api/v1/repositories?ownerId=…` 列出同一擁有者的來源。每頁先從操作者的候選集合讀取，再依擁有者、啟用狀態與即時權限過濾；空頁仍可能有下一頁。游標限原操作者與原查詢範圍使用。省略 `ownerId` 時沿用原本的跨擁有者查詢。

[上層導覽](../README.md)
