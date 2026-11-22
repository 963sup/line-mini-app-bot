# backend/internal/business/attendance 工作指引

## 用途與範圍

管理打卡、出勤紀錄、補卡、請假、審核與出勤看板。領域與應用層各依 [模組導覽](README.md#現行結構) 的職責套件分層。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[errors.go](application/port/errors.go)、[store.go](application/port/store.go)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：定位前預檢、狀態與時序](../../../../docs/domains/attendance/README.md#定位前預檢狀態與時序)

## 分層約束

本模組採用 domain/entity、valueobject、service、event、policy，以及 application/command、query、port、dto。新增 Go 檔必須放入對應職責套件，不能恢復舊的功能混合套件或用別名套件繞過拆分。依賴允許範圍由 `internal/architecture/boundaries_test.go` 檢查。領域層不持有查詢分頁與外部序列化標籤。`adapter/messaging`、`adapter/scheduler` 目前僅保留入口，不代表已實作。

## 群組打卡的維護與驗收

驗證同一人的不同群組班次分開維護，並覆蓋無法接受打卡的班次狀態。

修改此規則時，核對[判定流程](../../../../docs/domains/attendance/README.md#decision-flow)與[代表測試](adapter/firestore/attendance_integration_test.go)。
