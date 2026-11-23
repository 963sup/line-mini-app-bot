# 日誌、監控與追蹤

<a id="cloud-logging"></a>

## Cloud Logging

**狀態：營運基礎；採集與保存設定待核驗。**

**用途與最小流程**

以結構化日誌追查 API、LINE Webhook 與背景工作的成功、拒絕及失敗。

應用輸出結構化記錄 → 日誌依 service／revision／operation ID 搜尋 → 連回 trace 或告警。

**資料、授權與失敗邊界**

記錄動作、結果、錯誤類別與耗時；排除 token、Cookie、機密、原始訊息及不必要座標。

**導入條件**

指定保存期、查閱身分、環境與敏感欄位處理；日誌不取代業務決定的持久稽核。

**驗收**

可串起一次請求及後續工作，錯誤分類可查，敏感資料不外洩。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/logging/docs/overview)。[Monitoring](#cloud-monitoring) · [Trace](#cloud-trace)。

<a id="cloud-monitoring"></a>

## Cloud Monitoring

**狀態：營運基礎；指標與告警待核驗。**

**用途與最小流程**

觀測可用性、延遲、錯誤率、佇列積壓及處理失敗。

選取服務指標與業務結果 → 設定期間、門檻及告警 → 連結處置手冊。

**資料、授權與失敗邊界**

指標 label 使用有界維度，不放使用者 ID、原始 URL、機密或無限操作 ID。

**導入條件**

區分預期拒絕與服務故障，門檻依基線訂定；告警須有負責人、測試與解除條件。

**驗收**

模擬故障能通知、恢復可解除；無資料、重複告警及錯誤環境有處置。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/monitoring/docs/monitoring-overview)。[Logging](#cloud-logging) · [故障診斷](../../operations/troubleshooting.md)。

<a id="cloud-trace"></a>

## Cloud Trace

**狀態：按需導入。**

**用途與最小流程**

定位 HTTP、Firestore、LINE 與 AI 呼叫的延遲及失敗位置。

入口建立／接續 trace → 子呼叫建立 span → 背景工作以關聯資訊追查 → 匯出與日誌對照。

**資料、授權與失敗邊界**

僅傳遞必要 trace context；span 不記錄 token、訊息全文或敏感表單。背景工作是獨立執行，保留適當關聯。

**導入條件**

定義採樣率、匯出身分、逾時與成本；既有平台 trace 不代表應用內部呼叫已完整接線。

**驗收**

可找出慢速子呼叫，取消與失敗可辨識；關閉匯出不影響業務成功。

**來源與相關規格**

[官方文件](https://docs.cloud.google.com/trace/docs/overview)。[Logging](#cloud-logging) · [Monitoring](#cloud-monitoring)。

[Google 服務索引與查證範圍](README.md#sources)
