# backend/internal/business/summary 工作指引

## 閱讀與修改入口

先查閱[errors.go](application/port/errors.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：資料與可靠性](../../../../docs/domains/summary/summary.md#資料與可靠性)

## 驗證模型摘要的維護與驗收

更換模型後仍須檢查引用；結果缺少引用或引用超出來源範圍時，拒絕儲存與發送。

修改此規則時，核對[判定流程](../../../../docs/domains/summary/README.md#decision-flow)與[判定實作](application/summarization/validation.go)。
