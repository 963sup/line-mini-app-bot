# 身分與 Session 契約

## 用途與範圍

定義前端登入與身分相關的 API 契約。

## 架構與責任邊界

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[identity.yaml](identity.yaml)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[Session 設計](../../../docs/architecture/identity/browser-session.md) · [現行實作圖：續期交易與前端觸發](../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發)

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 維護 MINI App 公開設定、本人與 Session 建立、續期、撤銷的既有契約。
- 可信身分由後端驗證；客戶端輸入、登入或 profile 不等同 Repository 成員或其他業務操作權。
- 保留現有 Cookie／Bearer、CSRF、錯誤及單例路徑相容性；不在公開 DTO 暴露憑證與服務端機密。
- 修改前閱讀 [Session 設計](../../../docs/architecture/identity/browser-session.md)。

[現行實作圖：續期交易與前端觸發](../../../docs/architecture/identity/browser-session.md#續期交易與前端觸發) · [未來設計圖](../../../docs/diagrams.md#未來設計)
