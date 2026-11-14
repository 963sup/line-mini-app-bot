# 工程項目匯入草稿基礎

狀態：文字型 PDF 匯入最小切片已接上本機操作介面與單一 OpenAI adapter；本機以 fake provider 驗證，正式模型／LINE／手機環境尚待驗收。

本切片先在瀏覽器擷取文字型 PDF，使用者預覽、遮蔽並明確同意後，單一 OpenAI producer 只接收所選文字、產生不可信候選；驗證後才保存可人工審核的草稿，再經明確確認建立 Objects。原 PDF 不上傳或保存到伺服器；沒有 OCR、後端檔案庫或背景佇列。Objects 核心不依賴匯入或 AI。

## 文字 PDF 操作與設定

在儲存庫的 Objects／工程項目選擇 PDF 匯入。本機 PDF 最多10 MiB／50頁／128 KiB文字，擷取30秒截止；掃描無文字或密碼 PDF 拒絕，不執行文件 JavaScript、XFA、附件或外部 URI。PDF.js 6.5.438 legacy build 延遲載入，worker／CMap／fonts／WASM 同源自託管；不使用 CDN 或放寬 CSP。真實 LINE iOS／WKWebView 相容性仍需裝置驗收。

原始擷取文字唯讀留於記憶體；另有待傳送文字可遮蔽。所有頁面預設不選，建議先遮蔽不需分享的個資、金額及條款；核對內容可送出後，選明細頁並明確同意傳給 OpenAI。只發送已選取、已編輯文字；引用只能證明提交文字存在，不代表原 PDF 真偽。

伺服器沿用 `OPENAI_API_KEY` 環境變數／既有 Secret reference；不在瀏覽器放 key。`OBJECT_IMPORT_MODEL` 可選，預設 `gpt-4.1-mini`。缺 key 回明確未配置錯誤，本機缺值不代表程式未支援環境注入。解析模型80秒、路由85秒、單次瀏覽器請求90秒；全站逾時設定不改。

解析只產生候選，不寫草稿或 Object；失敗重試可能再呼叫模型。取得候選後固定完整草稿 body 與建立鍵，未知保存結果只重試相同 body／鍵，不重新呼叫模型。審查後若再編輯，必須重新保存人工審查才能確認；確認結果不明先 GET 進度，再沿原審查版本續傳。

瀏覽器只保存依身份／儲存庫隔離的草稿 ID 與原檔 SHA-256 續傳 metadata，不保存 PDF、原文、候選或憑證。重新整理後須重選同 hash 的本機檔案以恢復原文對照；localStorage 失敗時顯示草稿 ID 及無法保證恢復的提示，不宣稱持久化草稿原文。

## 輸入與驗證

POST 提供 sourceDocument、sourceRevision、pages（number／text）及候選 rows。每行包含 name、description、type、quantity、unit、必填 sourceItemNo、materialCode 與 evidence（page／quote）；不接受 actor、status、Object ID 或建立鍵。來源文字是不可信資料，不能當指令或授權。

- 最多 100 行、50 頁；頁號唯一且為正整數，原文 UTF-8 總上限 128 KiB，單頁 32 KiB；HTTP JSON 上限 512 KiB，拒絕未知欄位及額外 JSON。
- 每行 1–4 個 evidence；每個 quote 非空且最多 500 字元，必須是其指定頁原文的逐字子字串。這只證明引用存在於提交頁文，不證明 PDF 真偽或內容／數量的解讀正確。
- 欄位遵循 [Object 規則](objects.md)，類型是人工／producer 提案，quantity 不換算，狀態只能由 Objects 建立端初始化。原項次每行一筆，不依 quantity 拆分。trim 後完全相同的來源項次在同草稿拒絕（不推論 010 與 10 相同），不跨不同草稿自動去重。
- 首版只保存全部有效的候選；任何非法欄位／引用拒絕整次建立，回 400，不留下可確認的半草稿。持久化 v1 草稿 JSON 最多 256 KiB；輸入／review以投影所有結果 ID、完整 metadata 與最大版本後的最終快照，加 1024 bytes 餘量，採更保守的252 KiB預算，避免確認後才超限。

原始完整分頁文字只在建立驗證時使用，不保存於草稿；保存提交分頁文字內容雜湊（不是原 PDF 檔案真偽證明）、必要來源 metadata、短 quote／頁碼與候選。沒有把原始商業 PDF 放入 repo／fixtures／日誌，亦不自動傳給外部模型。

## 狀態與操作契約

| 欄位 | Create／Get | Review | Confirm／Retry |
| --- | --- | --- | --- |
| Actor／Resource | Session／Bearer 可信本人；指定 Repository 的 ObjectImport 草稿 | 同左；本人建立的草稿 | 同左；已人工 review 的固定快照 |
| Authority／Context | 目前成員、啟用容器；Get 僅建立者 | 目前成員且為建立者、expectedVersion、尚未 confirming | 同左；expectedVersion 必須指向本次 review 版本 |
| Policy／Decision | 驗證全部候選／來源，失敗不保存；不信任客戶端身份 | 只能修改既有各 row 的 Object 內容，不改 rowId／來源引用／文件；引用證據保持建立時已驗證快照；成功即明確人工審核 | reviewed 才可開始；交易凍結 confirmedFromReviewVersion，confirming 後不可再編輯；所有行驗證後才開始寫入 |
| Result | draft 狀態、server job／row ID、version=1 | 完整行內容、reviewed 狀態、版本遞增；過期版本 409 | 逐行呼叫唯一 object.Service.Create；每行 resultObjectId 與尚未完成狀態可由 GET 讀取（失敗確認回安全錯誤，不直接回完整草稿），全部完成才 completed |
| 重送／併發 | 建立必須 Idempotency-Key；同本人／repo／鍵與正規化候選及文件內容雜湊回原草稿，同鍵不同內容 409。不同鍵／新草稿可重複來源，不宣稱全文件去重 | 版本衝突先重讀 | server 由 jobId＋rowId 產生穩定建立鍵，保存逐行結果。未知提交／部分失敗或併發確認都沿用相同快照及鍵重試，不另生成鍵；相同 review 版本重試合法，其他版本 409 |
| Audit | 保存 owner／時間／來源雜湊，不記 raw text／token | 保存版本與 review 時間 | 保存確認版本與逐行 ID 映射；不修改已存在 Object、不通知外部服務 |

首版 Review 必須提交相同既有 row IDs，不支援增加／刪除行或修改 sourceItemNo／文件／引用；materialCode 可人工修正。最小介面不支援增刪列；候選不完整時，應在保存草稿前放棄候選、重新選頁解析。

Object 建立仍重新核對目前成員與啟用容器；匯入不能授權修改 Object 或覆寫新版 PDF。job 狀態交易與 Objects 建立是兩個交易，可能部分成功；不宣稱整批原子提交。建立成功但結果映射尚未保存時，重試相同鍵取回原 Object 後補映射。每行錯誤可安全重試，不將 SDK／provider 原文錯誤或來源全文回傳。

## API 與保存

權威契約位於 [object-imports.yaml](../../api/contract/object-import/object-imports.yaml)，使用既有驗證／CSRF／typed client 生成流程：

- POST `/api/v1/repositories/{repositoryId}/object-imports/parse` 接受來源、分頁文字及 `consent: true`，回有效候選與 sourceHash，不建立草稿／Object，也不承諾模型完整找出所有項次。
- POST `/api/v1/repositories/{repositoryId}/object-imports` 以 Idempotency-Key 建立有效候選草稿。
- GET `/api/v1/repositories/{repositoryId}/object-imports/{id}` 查看本人草稿及結果。
- POST `.../{id}/review` 提交完整人工審核行內容與 expectedVersion。
- POST `.../{id}/confirm` 以 review 的 expectedVersion 明確確認或重試。

草稿保存於 `channels/{channelId}/repositories/{repositoryId}/objectImports/{id}`，不另建帳號／角色／計數器。暫不配置自動 TTL／刪除，不宣稱已實體清除；保存的來源短引文與草稿仍應視為業務資料，不是公開內容。

## 驗收界線

以脫敏候選／fake HTTP provider 證明完整流程：模型不可信欄位、引用不存在、重複項次、大小上限、未登入／跨 repo／非 owner／停用、review 版本衝突、未 review 拒絕、凍結後不可改、部分成功與併發重試不重複、逐行結果可讀。使用真本機 Firestore emulator 驗交易；不以假 producer 宣稱 AI 已解析或正式 LINE／Firestore 已驗收。

[Object 規格](objects.md) · [功能索引](README.md) · [操作模型](../architecture/business-operation.md)
