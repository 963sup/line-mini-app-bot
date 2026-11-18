# 物件自動匯入

先選整批「施作／非施作」，再選文字型 PDF 後，瀏覽器擷取支援版型的明細，伺服器以單一 OpenAI 模型產生候選、驗證並自動保存為物件。結果可由既有物件詳情編修；正常流程不要求另存草稿、人工審查或確認建立。這是受控工作流程，模型不能直接操作資料庫或決定權限。

## 使用與資料邊界

在儲存庫「物件」先選「整批類型」（施作／非施作），再選「選擇PDF並匯入」。預選施作可見且可改，整批先套此類型，匯入後可逐筆編修；不是模型分類或server替使用者選類型。選檔處明示會將明細文字經伺服器傳送至 OpenAI；使用者選檔才啟動處理，不另外要求兩個勾選或手填來源／版本。

- 原 PDF 僅在瀏覽器處理，不上傳或永久保存全文。只支援可辨識明細表頭與結尾的文字訂單；本機依座標保留項次、品名／說明、數量與單位，排除右側價格欄、頁首與表外條款。未知版型明確拒絕，不退回外傳全文；不支援掃描 OCR、密碼 PDF 或任意文件格式。
- 檔案最多 10 MiB、50 頁，明細文字合計 128 KiB、單頁 32 KiB；本機擷取有 30 秒截止。PDF.js legacy build 與 worker／字型／CMap／WASM 同源自託管，不執行 PDF JavaScript、XFA、附件或外部 URI。
- 類型隨原請求body／key凍結，抽取、未知結果或部分成功期間不能更換；來源識別自動使用原件雜湊，版次僅取文件內明示值，不以檔名或雜湊冒充商務版本。不確定版次留空；模型只收到必要明細頁文，不需要訂單號或檔名。
- 新自動路徑的AI只回傳name、quantity、unit及安全追溯用sourceItemNo／pages，不回傳type、description、materialCode；不要求模型重抄長引文作為入庫門檻。頁參照與欄位有效不代表模型解讀或原件真實性已認證，AI仍可能漏項或誤解，使用者可檢查結果並編修；completed只表示已凍結候選全部建立，不認證原件所有項次皆已辨識。

伺服器沿用 `OPENAI_API_KEY` 與既有 Secret reference；`OBJECT_IMPORT_MODEL` 可選，預設 `gpt-4.1-mini`。金鑰不進瀏覽器。本機替身與測試不能證明正式模型、LINE 或手機環境已驗收。

## 自動匯入操作契約

權威 HTTP 契約為 [object-imports.yaml](../../../api/contract/object-import/object-imports.yaml)。新增 POST `/api/v1/repositories/{repositoryId}/object-import-runs` 與 GET `.../{id}`，舊 API 語意保持相容。

| 欄位 | 自動匯入／重試 | 查看結果 |
| --- | --- | --- |
| Actor／Resource／Action | Session／Bearer 可信本人；指定 Repository 的匯入執行；解析並建立物件 | 同本人、同儲存庫的執行結果 |
| Context／Authority | 當下群組成員、啟用容器；Idempotency-Key 與提交來源內容 | 當下成員、啟用容器且為執行建立者 |
| Policy／Decision | 驗證全部候選後才凍結；模型不能指定 actor、狀態、ID 或建立鍵；每項次一筆，不依數量拆分 | 不揭露其他本人／儲存庫的執行 |
| Result | pending 尚無候選快照；processing 已凍結候選且建立中；completed 全部結果映射完成。安全 reason 說明可重試失敗 | 回傳保存快照、每列已建立 ID 與尚未完成結果 |
| 重送／併發 | 同本人／儲存庫／鍵與內容（含所選type）回同一執行；同鍵異內容或異類型409。只有首個候選快照可凍結，競爭者必須沿用贏家的列與鍵 | 讀取不重新解析 |
| Audit／保存 | 保存來源指紋、建立者／時間、候選與逐列結果；不保存原 PDF、整份頁文、金鑰或模型原始回覆 | 不額外記錄原文或個資 |

候選遵循 [物件欄位規則](objects.md)：正十進位數量字串、原單位不換算、類型只能由使用者此次選擇灌入、同執行項次不可重複。施作初始完成數0與not_started僅由後端物件建立服務設定；非施作不給完成數，畫面狀態／進度不適用。AI不再依費用、章節或內容推斷類型；新列description／materialCode留空，不刪改舊凍結列的既有資料。舊人工草稿自由type保持相容。模型或輸入失敗不建立無效物件；不同建立鍵仍可能重複同份文件，不宣稱全文件去重。

模型呼叫在交易外。候選凍結後先以保守容量預算確認所有結果 ID 可保存，再逐列呼叫唯一 `object.Service.Create`；每列使用穩定建立鍵及結果映射。已凍結執行不再解析，部分成功或回覆遺失只續跑同一快照，不覆寫既有物件。不是整批原子提交；畫面必須如實呈現已建立與待重試數量。

瀏覽器僅在目前頁面記憶體保留未確認請求 body／鍵；不保存原文至 localStorage。回覆不明時請在原頁重試，不宣稱重新整理後可恢復原意圖。首次明確回傳 pending、尚無候選及物件且從未結果不明時，可更換PDF；已建立部分結果或曾結果不明則沿用原意圖續跑。

當下成員資格由外部服務觀察，無法與 Firestore 形成原子快照；交易仍核對容器啟用與群組綁定。不能以畫面或模型輸出取代後端授權。

## 舊自動執行相容

creation.type在公開schema維持optional以接收舊body，但新意圖必須明確提供work／non_work；缺值在任何新保存／模型呼叫前拒絕，不能暗中預設。run ID方向不變，新意圖指紋v2綁定所選type。

舊v1已凍結的processing／completed執行，原本缺type的相同source指紋仍能沿舊列續跑，不再解析、不改類型；帶新type套舊key視為不同意圖409。舊pending尚未指定類型時明確提示重新選擇並建立新意圖，保留舊資料，不猜類型、不宣稱其他舊請求必定未提交，也不擴張跨key文件去重承諾。

## 相容的人工草稿 API

既有 `/object-imports/parse` 仍只產生候選，不改成寫入。`/object-imports`、`/{id}`、`/{id}/review`、`/{id}/confirm` 仍保留人工草稿／逐字引文／版本確認與原有儲存格式，供既有客戶端相容；新畫面不偽造人工審查紀錄，也不要求新流程經過這些步驟。

舊草稿存於 `channels/{channelId}/repositories/{repositoryId}/objectImports/{id}`；新自動執行存於同儲存庫的 `objectImportRuns/{id}`。不自動刪除舊資料或配置 TTL，不宣稱業務資料已實體清除。

## 驗收界線

以脫敏資料與 fake HTTP provider 驗證候選有效性、權限／跨儲存庫、同鍵內容衝突、候選凍結競爭、部分成功與未知結果重試；以真本機 Firestore emulator 驗交易與結果映射。真 PDF 只做本機擷取驗證，不將商業內容放入 repo／公開 fixture 或傳至模型。正式 OpenAI、LINE 與手機驗收另列，不能用替身、emulator 或 build 冒稱通過。

[物件規格](objects.md) · [功能索引](../README.md) · [操作模型](../../foundation/architecture/business-operation.md)
