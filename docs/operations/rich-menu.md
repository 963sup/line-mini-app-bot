# Rich Menu 自動發布

Rich Menu 由部署流程自動同步，不需人工發布。程式入口為 server 的 `sync-line-navigation` 工作模式；原 `backend/cmd/richmenu` 與映像 `/richmenu` 已移除。一般 HTTP 服務啟動不會操作選單。

## 流程與分工

[部署設定](../../deploy/cloudbuild.yaml) 在映像推送後建立每個 Build 專用的 Cloud Run Job，等待選單同步成功後才部署 HTTP 服務，成功後移除 Job。失敗 Job 保留供診斷。未改變的選單不寫入 LINE；Alias 切換不是原子操作，失敗重跑會收斂。

[導航](../../backend/internal/transport/line/navigation/menu.go) 定義內容與 attendance-in/out Alias；[平台機制](../../backend/internal/platform/line/richmenu/README.md) 負責圖片、發布、Alias、預設選單與安全清理；[bootstrap](../../backend/internal/bootstrap/line_navigation.go) 讀取所選 MINI App 設定並注入依賴。圖片維持 assets/line/richmenu/*.jpg，映像內為 /assets/line/richmenu/，大小為 2500×843、最多 1 MiB。

## 雲端接線與前置條件

現有 managed Trigger 使用 inline build，修改 repository YAML 不會更新它。**部署這版映像之前，必須將該 Trigger 改用 repository 的 deploy/cloudbuild.yaml，或同步其中的 Job 與清理步驟，移除舊 /richmenu 呼叫。** 舊 CLI 參數不再支援。需保留實際 Trigger 的專案、區域、服務帳戶、substitutions 與既有部署設定，不能直接用範例值覆蓋。

2026-11-21 嘗試唯讀查詢 Trigger 時，現有 Google 憑證失效，無法核對或修改線上設定。本次僅完成程式與 repository 部署接線，未啟用線上流程或執行真實 LINE 寫入。

Job 使用既有 runtime service account，透過 Secret Manager 綁定 Messaging API token；需有該 Secret 存取權及指定 Firestore database 的讀寫權。Build service account 需有管理、執行 Cloud Run Job 與使用 runtime service account 的權限；未擅自變更 IAM。Job 使用所選 MINI App 三元組與 Messaging API Channel ID，不接受人工指定不同的 App URL。

## 互斥、重試與故障處理

Firestore 的 line_navigation_publication_leases/{channelID} 保存暫時 owner 與 expiresAt，與業務資料分開。Job 最多運行八分鐘，租約十分鐘，LINE 請求共用較短 context；取得租約前不發布，等待也計入工作時限。釋放時核對 owner，避免刪除別人的租約。程序崩潰後須待租約過期才能接手。此租約假設執行環境時鐘同步，不是 LINE API 支援的 fencing token。

單次工作最多重試一次，每次皆重新核對遠端內容。圖片失敗不修改 live Alias；Alias 部分更新會保留已準備完成的目標，下一次同步收斂。只有成功遷移使用者後才刪除舊 managed 選單。同步完成後 HTTP 部署失敗不會自動回滾選單；不同 Build 仍可能先後發布不同版本，租約只保證同時一個發布者，不保證提交順序。

## 驗收與來源

本機平台與 bootstrap race 測試、Architecture Dependency Boundaries、Lint 與編譯通過；Go Browser 測試依賴前端靜態產物，須先執行前端 build，再執行全庫 check；不以手寫 HTML 補缺。Firestore 租約案例需 FIRESTORE_EMULATOR_HOST；本環境 emulator 需要 Java 25 而僅有 Java 21，故該案例跳過，未宣稱通過。Mermaid CLI 未安裝，更新的既有目錄圖文字未重新渲染。部署後另核對 Job 成功、兩個 Alias、預設選單、圖片與手機切換；mock 不代替正式渠道驗收。

LINE SDK 維持 v8.23.0。官方 [Rich Menu API](https://developers.line.biz/en/reference/messaging-api/#rich-menu) 與批次遷移規範核對自 line-developers-docs-source main / c15d6d3fa2eff7fcae65554c20be5b1e60f9a10a。Firestore transaction 語意核對 googleapis/google-cloud-go firestore/v1.29.2 原碼文件；交易 callback 可能重跑，因此只取得／釋放租約，所有 LINE 副作用均在交易外。Context7 配額已耗盡；主機未提供另一條匿名連線。gcloud Job 旗標以本機固定版 592.0.0 leaf help 核對；未做雲端整合驗收。

[LINE 接入入口](../integrations/line/setup.md) · [Cloud Run 部署](cloud-run.md)
