# 共用傳輸契約

## 用途與範圍

保存各領域 API 共用的 HTTP 傳輸契約。

## 架構與責任邊界

維護共用的 HTTP 傳輸定義，例如認證與錯誤回應。領域操作與業務型別由各能力維護。

## 閱讀與修改入口

先查閱[transport.yaml](transport.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../../docs/architecture/contracts/http.md#基礎入口)

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 僅管理真正共用的錯誤格式、認證方式及必要傳輸定義；不收納 Repository、Issue、Discussion、Project 或 Attendance 的業務模型。
- 共用認證 scheme 只描述憑證形式，不代表取得群組、寫入或審核權限；各操作仍須定義自己的授權與失敗語義。
- 修改共用定義須核對所有引用者及相容性，不依賴任何業務能力或組合入口。

[現行實作圖：基礎入口](../../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../../docs/diagrams.md#未來設計)
