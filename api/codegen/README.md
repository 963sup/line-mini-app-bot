# api/codegen 目錄導覽

## 用途與範圍

配置公開與內部 Go API 的程式碼產生流程。

## 架構與責任邊界

契約是介面來源；client 與產生設定供消費端使用，依契約重新產生型別。

## 閱讀與修改入口

先查閱[go.yaml](go.yaml)、[internal-go.yaml](internal-go.yaml)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../docs/architecture/contracts/http.md#基礎入口)

從 [go.yaml](go.yaml)、[internal-go.yaml](internal-go.yaml) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：基礎入口](../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../docs/diagrams.md#未來設計)
