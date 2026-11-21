# assets 工作指引

## 閱讀與修改入口

先查閱[line/richmenu](line/richmenu)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[文件索引](../docs/README.md)


適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

line/richmenu/ 保存兩張選單 JPEG 與既有操作指南。圖片由容器打包，對應 backend/cmd/richmenu CLI。

修改圖片時核對既有尺寸、點擊區域與用途；素材變更不等於 LINE 選單已更新。遠端發布須有明確授權。

## 驗證

閱讀 line/richmenu/README.md 並核對圖片尺寸與 Dockerfile 素材路徑；不以 Graphify 未收錄圖片推論素材不存在。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。
