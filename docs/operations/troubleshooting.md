# 疑難排解

## 診斷順序

1. 記錄時間、環境、revision、入口、actor／resource 範圍與預期結果。
2. 重現最小案例，區分客戶端、契約、後端、資料庫與外部供應端。
3. 以 correlation ID 串起 request、transaction、task 與外部呼叫。
4. 檢查最近部署、設定、Secret reference、IAM、索引與資料版本。
5. 確認失敗是否可重試，以及重試是否需沿用 operation ID。

## 常見分類

| 症狀           | 優先檢查                                                |
| -------------- | ------------------------------------------------------- |
| 401／403       | token／Session、Origin／CSRF、audience、資源授權        |
| 409            | 資源版本、idempotency key、狀態轉換                     |
| 5xx            | 下游分類、timeout、Firestore、設定缺失                  |
| LINE 無回覆    | Webhook 簽章、事件處理、reply token、重送               |
| 前端顯示舊資料 | 身分世代、取消、遲到回應、BFCache                       |
| 部署後異常     | revision、traffic、runtime env、secret、service account |

## 安全處理

排查過程不輸出 token、cookie、secret payload 或使用者內容。需要檢查憑證時，只驗證能否取得或使用，將敏感 stdout 丟棄。正式資料修復、重新部署、IAM 或流量變更都需要獨立授權。

## 完成條件

修復後重跑原始案例與相鄰拒絕案例，說明根因、必要修改、驗證結果及尚未覆蓋的環境。只找到暫時 workaround 時，保留根因與移除條件。

## Firestore 唯讀連線檢查

設定 `GOOGLE_CLOUD_PROJECT`；`FIRESTORE_DATABASE_ID` 預設為 `(default)`。本機使用應用程式預設憑證（ADC），Cloud Run 使用執行服務帳戶。從儲存庫根目錄執行，將 `COLLECTION/DOCUMENT` 換成一份已存在的文件：

```bash
go -C backend run . check-firestore -document COLLECTION/DOCUMENT
```

[檢查程式](../../backend/internal/bootstrap/firestore_check.go)對指定文件執行一次 `Get`。成功代表目前身分可讀該文件；程式不印出文件內容，也不寫入資料。若收到 `NotFound`，請核對資料庫和文件路徑，讀取權限仍待確認。

## Rich Menu 發布互斥與重試

Firestore 的 line_navigation_publication_leases/{channelID} 保存暫時 owner 與 expiresAt，與業務資料分開。Job 最多運行八分鐘，租約十分鐘，LINE 請求共用較短 context；取得租約前不發布，等待也計入工作時限。釋放時核對 owner，避免刪除別人的租約。程序崩潰後須待租約過期才能接手。此租約假設執行環境時鐘同步，不是 LINE API 支援的 fencing token。

同步依選單定義與圖片的 SHA-256 判定是否需要寫入，並讀取遠端圖片、Alias 與預設選單以偵測外部變更。租約仍用於互斥；重複同步行為由[發布測試](../../backend/internal/platform/line/richmenu/publication_test.go)驗證。

單次工作最多重試一次，每次皆重新核對遠端內容。圖片失敗不修改 live Alias；Alias 部分更新會保留已準備完成的目標，下一次同步收斂。只有成功遷移使用者後才刪除舊 managed 選單。同步完成後 HTTP 部署失敗不會自動回滾選單；不同 Build 仍可能先後發布不同版本，租約只保證同時一個發布者，不保證提交順序。

## Rich Menu 舊選單遷移

`rich-menu-clock-in`、`rich-menu-support` 與 `rich-menu-clock-out` 是正式出勤選單；`attendance-in`、`attendance-out`、`rich-menu-transfer` 僅保留為既有選單的舊 Alias。發布前核對已存在的舊 Alias 確實指向本流程管理的選單。新圖與新 Alias 就緒後，將已存在的舊 Alias 保留為轉向入口，並讀回確認。缺少的舊 Alias 不新增。

圖片由根目錄執行 `pnpm rich-menu:generate TC_FONT_OTF_OR_TTF` 重新產生。腳本以 `rich_menu.json`、MINI App 共用主題 Token 與指定繁中字型為輸入，直接覆寫 `assets/line/richmenu/` 的三張 PNG；Go 使用同一份 JSON 組裝 LINE SDK 點擊區域。上方在上班、支援、下班情境間切換，中間執行當前情境的唯一主要操作；底部左側返回個人功能選單，右側開啟出勤 MINI App。當前情境不建立無效的自我切換區域。

已綁定個別使用者的舊選單須另外透過 LINE 批次遷移；只有遷移成功且無 Alias 仍引用舊選單才刪除。失敗保留舊選單，重跑同步可繼續。驗收須在測試渠道涵蓋仍綁舊頁的使用者，並保留管理權衝突時的停止行為。
