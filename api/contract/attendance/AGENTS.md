# 出勤契約

## 用途與範圍

定義打卡、出勤紀錄、工時、請假、補卡及審核的 API 契約。

## 架構與責任邊界

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[approvals.yaml](approvals.yaml)、[corrections.yaml](corrections.yaml)、[leave.yaml](leave.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[出勤](../../../docs/domains/attendance/README.md) · [查詢與更正](../../../docs/domains/attendance/queries.md)

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 維護打卡、查詢、工時、更正、審核與請假的既有操作；按 [能力索引](../README.md) 保持型別擁有者及依賴方向。
- 修改前閱讀對應功能規格：[出勤](../../../docs/domains/attendance/README.md)、[查詢與更正](../../../docs/domains/attendance/queries.md)、[工時](../../../docs/domains/attendance/worktime.md)、[請假](../../../docs/domains/attendance/leave.md)。
- 保留既有 JSON 大小寫、時間與月份、nullable、分頁、請求上限及 Idempotency-Key 語義；不藉命名整理更改公開路徑。
- 登入、Repository 存取與審核權分開；後端執行本人範圍、群組權、版本與交易檢查。契約驗證不替代業務授權。
- 請假登記不代表已有請假審核；不把未實作能力列為可用操作。

[現行實作圖：定位前預檢、狀態與時序](../../../docs/domains/attendance/README.md#定位前預檢狀態與時序) · [未來設計圖](../../../docs/diagrams.md#未來設計)
