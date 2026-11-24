# 可觀測性

可觀測性用於回答哪個操作、在哪個版本、經過哪些邊界、以什麼結果結束。

## 訊號

- Log：結構化事件、穩定 severity、operation／correlation ID、bounded error class。
- Metric：請求量、延遲、失敗、重試、佇列與資源飽和；label 不使用個別 user 或 resource ID。
- Trace：HTTP、gRPC 與外部依賴的跨邊界時序。

Go service 使用 `log/slog` 與 OpenTelemetry；Google Cloud exporter 與實際啟用狀態以 bootstrap 及目標環境為準。

## 記錄邊界

不得記錄 access token、ID token、cookie、CSRF、secret、完整 LINE payload、文件內容或不必要個資。供應端原始錯誤先分類與遮罩，再記錄安全摘要。資源追查使用不可逆或受控 reference。

## 操作結果

重要操作至少能區分 success、rejected、conflict、retryable failure、permanent failure 與 unknown outcome。Webhook 接收、資料提交與外部送達是不同階段，metric 與 log 不應混成單一成功。

## 告警

告警需有 owner、門檻、觀察窗口與 runbook。優先涵蓋不可用、錯誤率、延遲、工作積壓、認證異常與部署回歸；避免只因單次錯誤或高基數 label 觸發。
