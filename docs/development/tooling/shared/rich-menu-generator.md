# Rich Menu 元件化與產生器設計

全高版素材、產生器與發布接線已實作。產生器從同一份選單定義輸出三張 **2500 × 1686 px JPG**，同時提供發布所需的 LINE 熱區與 Action。發布端在同步前核對 JPG 尺寸與請求尺寸。發布操作見 [Rich Menu 自動發布](../../../operations/rich-menu.md)，任務行為見 [LINE 任務接線表](../../../integrations/line/wiring.md)。

## 版面與按鈕

畫布、格線、圖片與操作對應見[Rich Menu 版面](../frontend/rich-menu-layout.md#版面與按鈕)。

## 共用定義與產生原理

現行實作

```mermaid
flowchart LR
  S["單一 JSON：尺寸、格線、頁面、Action"] --> R["本機產圖：共用分頁與主操作元件"]
  S --> L["Go 導覽：LINE 尺寸、熱區、Action"]
  R --> A["三張 JPG：可檢查是否過期"]
  A --> J["既有發布 Job"]
  L --> J
```

[共用定義](../../../../backend/internal/transport/line/navigation/attendance-rich-menu.json)同時供 [Go 導覽](../../../../backend/internal/transport/line/navigation/rich_menu.go)與 [圖片產生器](../../../../scripts/generate/rich-menu.py)讀取。分頁位置由欄寬累加，主操作從分頁高度以下開始。同一頁的 alias、主文案與 Postback 都在同一筆定義。第一頁是預設頁。

## 本機操作

使用具備 Pillow 13.0.0 的 Python，以及 Noto Sans CJK 字型集合 `NotoSansCJK-Regular.ttc`。預設字型位置為 `/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`。其他位置可用 `--font` 指定。程式要求字型 SHA-256 與專案使用版本一致，避免相同文案換字型後造成版面漂移。JPEG 位元組比對使用相同的 FreeType／JPEG 環境。

```sh
python3 scripts/generate/rich-menu.py generate
python3 scripts/generate/rich-menu.py check
```

`generate` 先渲染並驗證全部圖片，再替換素材。`check` 在記憶體重畫，若任一檔案缺少或內容過期，列出檔名並退出失敗。可使用 `--output-dir` 在暫存目錄預覽或驗證。修改定義後，先跑 `generate` 和 `check`，再跑導覽測試，確認只對其他分頁配置切換熱區，中央 Postback 有可用的 handler。

正式發布流程仍讀取提交的 JPG。生成器不呼叫 LINE 或部署；正式預設頁、三頁切換和私訊指引須在手機另行驗收。[LINE 圖片規格](https://developers.line.biz/en/reference/messaging-api/#upload-rich-menu-image)
