# 應用素材

## 用途與範圍

保存應用程式與 LINE 介面使用的資產。

## 架構與責任邊界

資產在此維護，功能行為由前後端實作。

## 閱讀與修改入口

先查閱[line/richmenu](line/richmenu)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[本層規格與指引](../docs/guides/line-setup/README.md)


## 目錄導覽與維護

line/richmenu/ 保存兩張選單 JPEG 與既有操作指南。圖片由容器打包，對應 backend/cmd/richmenu CLI。

修改圖片時核對既有尺寸、點擊區域與用途；素材變更不等於 LINE 選單已更新。遠端發布須有明確授權。

閱讀 line/richmenu/README.md 並核對圖片尺寸與 Dockerfile 素材路徑；不以 Graphify 未收錄圖片推論素材不存在。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。
