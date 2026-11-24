# 環境與整合驗收

[ADR 0002](../../decisions/0002-baseline.md)保留早期出勤基線與交付順序；現行功能以[領域規格](../../domains/)為準。依本次變更及目標環境選擇驗收，不將歷史交付階段視為每項任務的執行順序。

- **開發環境與容器：** 工具與服務依[環境指南](codex-cloud.md)、[工具鏈](../tooling/shared/mise.md)及[容器驗收](../../operations/container-build.md)處理。既有快取不能證明乾淨環境可重建。
- **Cloud Run：** 部署須確認授權、目標、服務身分、IAM、機密、觸發設定與實際 revision；依[發布驗收](../../operations/release-acceptance.md)記錄結果。
- **Firestore 與 LINE：** 依[Firestore 整合](../../integrations/firestore/connection.md)和 [LINE 整合](../../integrations/line/setup.md)核對身分、資料庫及權限；日誌不得記錄 token。真實收發、登入與 Rich Menu 發布各需對應證據。
- **出勤：** 功能與權限依[出勤規格](../../domains/attendance/design.md)驗收；Firestore 交易與真實手機行為分別記錄證據。

驗證範圍與未完成項目依[驗收矩陣](../tooling/testing/commands.md)和[證據限制](../verification/evidence.md)記錄。
