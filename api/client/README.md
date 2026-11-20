# api/client 目錄導覽

## 用途與範圍

封裝前端使用的 TypeScript API client 與公開型別。

## 架構與責任邊界

契約是介面來源；client 與產生設定供消費端使用，依契約重新產生型別。

## 閱讀與修改入口

先查閱[package.json](package.json)、[tsconfig.json](tsconfig.json)、[turbo.json](turbo.json)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../docs/architecture/contracts/http.md#基礎入口)

從 [package.json](package.json)、[src](src)、[tsconfig.json](tsconfig.json)、[turbo.json](turbo.json) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：基礎入口](../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
