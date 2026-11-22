# api/codegen 目錄導覽

配置公開與內部 Go API 的程式碼產生流程。

契約是介面來源；client 與產生設定供消費端使用，依契約重新產生型別。

## 閱讀與修改入口

從 [Redocly lint／bundle 設定](redocly.yaml)、[go.yaml](go.yaml)、[internal-go.yaml](internal-go.yaml) 開始。修改前讀[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：基礎入口](../../docs/architecture/contracts/http.md#基礎入口)

[上層導覽](../README.md)
