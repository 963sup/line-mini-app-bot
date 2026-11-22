# backend/internal/business/discussion 工作指引

## 閱讀與修改入口

先查閱[應用服務](application/service.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：最小實作邊界與驗證](../../../../docs/domains/repository/discussion/design.md#最小實作邊界與驗證)

## 調整討論分類的維護與驗收

只具分類權限的成員可以調整分類；同一請求若夾帶內容修改，應拒絕越權變更。

修改此規則時，核對[判定流程](../../../../docs/domains/repository/discussion/README.md#decision-flow)與[代表測試](domain/entity/classification_test.go)。
