# 基礎群聊摘要

實作範圍：群組文字收集、按需摘要與開關；沒有排程摘要、私訊群聊、查姓名、向量搜尋或管理後台。
這是程式交付說明，不代表 OpenAI、Gemini、Cloud Tasks、Firestore 或 LINE 已完成真實整合驗收。

## 使用

完成部署設定後，Bot 所在群組預設開啟，沒有加入或啟用公告。一般文字保存時不回覆，也不呼叫模型。

| 指令 | 行為 |
| --- | --- |
| 摘要 | 整理最近24小時、最多最近100則文字，結果 Push 回原群組。 |
| 關閉摘要 | 立即停止收集／使用舊資料，背景分批刪除舊訊息與摘要工作。 |
| 開啟摘要 | 恢復收集，開啟新的資料版本，不恢復舊資料。 |

各群組成員皆可操作，不需人工核准。開關指令會回覆操作結果；不是主動公告。
每則記錄 message ID、event ID、發言 user ID、文字與事件時間。不查 profile、不保存真實姓名；傳給模型時使用成員代號。
不收集私訊、位置、圖片、貼圖、出勤與其他既有指令；沒有 Bot 加入前的歷史訊息。
摘要輸出討論重點、決定與待辦、未確認說法；摘要不是事實查證。引用 [S1] 對應工作內 Sources 第一個訊息 ID。
第一版沒有公開來源查詢指令，也不聲稱能直接跳轉 LINE 原始訊息。

## 部署設定（不自動建立雲端資源）

沿用 GOOGLE_CLOUD_PROJECT、LINE_MESSAGING_API_CHANNEL_ID、LINE_MESSAGING_API_CHANNEL_SECRET、LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN、FIRESTORE_DATABASE_ID 與 runtime ADC。
以下 provider、選用的 API key、模型與三項 task 設定全部未設定時，摘要模組不啟用；設定任何一項後則檢查完整性。這是模組部署條件，群組使用時無需再逐一開啟。

| 環境變數 | 值 |
| --- | --- |
| SUMMARY_PROVIDER | `openai` 或 `gemini`；未設定時沿用 Gemini。 |
| OPENAI_API_KEY | 選用 OpenAI 時，由 Secret Manager 的 `openai-api-key` 注入。 |
| GEMINI_API_KEY | 選用 Gemini 時，由 Secret Manager 注入。 |
| SUMMARY_MODEL | 該 key／專案可使用、支援對應 provider API 的文字模型名稱（OpenAI 使用 Chat Completions、Gemini 使用 generateContent）；不硬編碼版本。 |
| SUMMARY_TASK_QUEUE | projects/PROJECT_ID/locations/REGION/queues/QUEUE_ID。 |
| SUMMARY_TASK_URL | 本服務 https://HOST/internal/summary，不能有 query 或 fragment。 |
| SUMMARY_TASK_SERVICE_ACCOUNT | Cloud Tasks 發送 OIDC token 使用的專用服務帳戶 email。 |

需事先啟用 Cloud Tasks API、建立 HTTP queue。建議 queue 設定 maxAttempts=5、minBackoff=120秒、maxRetryDuration=3600秒，避免無限模型重試。
runtime 帳戶需指定資料庫讀寫、指定 queue enqueue 權限，以及代用指定 task service account 的權限；Cloud Tasks service agent 需能產生該身分的 OIDC token。
task service account 與 queue 使用同專案。入口驗證 Google token 簽章、有效期限、audience（完整 SUMMARY_TASK_URL）、email 與 email_verified。
現有服務公開模式不能保護此內部入口，所以不能移除應用層 token 檢查。若服務另外啟用 Cloud Run invoker IAM，task service account 也需該服務 invoker 權限。
Cloud Run 請求逾時需至少100秒；程式 worker deadline80秒、模型45秒、LINE Push10秒，租約120秒。仍是同一個容器，沒有常駐 goroutine worker。
LINE Console 開啟 Webhook redelivery；所選模型 API key 的計費與模型資料處理政策應依實際帳戶確認。

截至 2026-11-08，serving revision `line-mini-app-bot-00027-695` 設定 OpenAI／
`gpt-4.1-mini`，以 Secret Manager `openai-api-key:1` 注入；Cloud Tasks queue 為 RUNNING，
runtime service account 在該 queue 具有 enqueue 權限。`gemini-api-key` 尚無任何 version，
因此 `GEMINI_API_KEY` 未綁定。這些 metadata 與 revision Ready 只證明初始化成功；尚未以
真實群組事件證明 Cloud Tasks OIDC、模型回應與 LINE Push 的完整鏈路。

## 資料與可靠性

資料以 channels/{channelID}/summaryGroups/{hash(groupID)} 隔離：

- 群組文件：Disabled、Epoch、Revision、LastControl；缺少文件表示預設開啟。
- summaryMessages/{hash(messageID)}：Message 欄位。OccurredAt／ExpireAt 為 UTC timestamp，內容可讀期24小時。
- summaryJobs/{hash(eventID)}：Job 欄位。Sources 僅存訊息 ID，不複製原文；Result 僅暫存以供 Push 重試，工作結束清空 Result。

請針對 summaryMessages、summaryJobs collection group 的 **ExpireAt** 欄位設定 Firestore TTL，並停用 Text、Result 的不必要索引。
TTL 可能延遲刪除，程式讀取仍會排除過期資料。工作期限也受最早來源到期時間限制，避免重試沿用過期內容。
一般查詢只按 OccurredAt 範圍／倒序限制100筆，再按時間正序送模型；每次模型輸入最多約24KB JSON，超出時僅採可容納的部分並標示採用筆數。
保存來源 user ID 不等於驗證真實身分，也不代表發言為事實。

訊息以穩定 message ID 原子建立；命令以 event ID 原子建立工作；Cloud Tasks 也使用固定 task ID。佇列建立失敗回傳503，由 LINE 重送恢復。
任務先取得租約，生成結果後先保存，再用固定 UUID retry key Push。Push 回應409視為已接受，其他錯誤重試時沿用結果。
模型呼叫完成但結果尚未保存時若程序中斷，可能重新呼叫模型；不宣稱模型費用 exactly-once。不同摘要指令是不同工作，第一版沒有跨請求快取或頻率限制。

收到收回事件時覆寫為不含原文的 tombstone（保留48小時防止延遲原訊息恢復），使舊版本工作失效並清除相關結果；舊模型結果不能寫回。
關閉時版本立即切換，清理 task 刪除舊版本子集合文件，並不假設刪除父文件可連帶刪除。
發送前重查版本；已開始的外部 Push、已送出的 LINE 訊息或已提交給模型的請求無法藉由後續收回撤銷。

## 驗證與來源

測試涵蓋群組文字靜默收集與發言者 ID、指令不收集、命令重送、輸入預算與成員代號、生成時關閉、Push 失敗沿用結果、模型截斷／虛構引用拒絕。
未以 mock 結果宣稱 Firestore 交易、OIDC、真實模型或 LINE 送達已驗收；啟用前需完成 staging 整合測試。

- [LINE receiving messages](https://developers.line.biz/en/docs/messaging-api/receiving-messages/)（官方文件來源 revision 090500a）。
- [Gemini generateContent](https://ai.google.dev/gemini-api/docs/text-generation)（Context7 官方文件查詢，模型名稱由設定指定）。
- [Cloud Tasks create](https://docs.cloud.google.com/tasks/docs/create-tasks)、[task de-duplication](https://docs.cloud.google.com/tasks/docs/reference/rest/v2/projects.locations.queues.tasks/create)。
- [Firestore TTL](https://docs.cloud.google.com/firestore/native/docs/ttl)。Developer Knowledge MCP 未連線，本輪使用 Context7 取得官方文件。

OpenAI adapter 使用 `github.com/sashabaranov/go-openai`（版本以 go.mod 為準）。兩個 provider 共用匿名化輸入、指令與引用檢查；截斷、拒絕、空結果或無效引用不發送。沿用現有 Cloud Tasks 重試，不在模型 adapter 另外重試。部署預設 OpenAI／`gpt-4.1-mini`，可用 substitutions 切回 Gemini 並同步設定對應模型。模型存取權、Secret 版本及實際呼叫仍需線上驗收。
