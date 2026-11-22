# Cloud Trace 簡要規格

**狀態：按需導入。** 本頁規定本專案的使用邊界；雲端是否已啟用、費率、配額及實際部署另行驗證。

## 用途與最小流程

定位 HTTP、Firestore、LINE 與 AI 呼叫的延遲及失敗位置。

入口建立／接續 trace → 子呼叫建立 span → 背景工作以關聯資訊追查 → 匯出與日誌對照。

## 資料、授權與失敗邊界

僅傳遞必要 trace context；span 不記錄 token、訊息全文或敏感表單。背景工作是獨立執行，保留適當關聯。

## 導入條件

定義採樣率、匯出身分、逾時與成本；既有平台 trace 不代表應用內部呼叫已完整接線。

## 驗收

可找出慢速子呼叫，取消與失敗可辨識；關閉匯出不影響業務成功。

## 來源與相關規格

[官方文件](https://docs.cloud.google.com/trace/docs/overview)，查證範圍見[整合總覽](README.md#sources)。[Logging](cloud-logging.md) · [Monitoring](cloud-monitoring.md)。

[Google 服務索引](README.md)
