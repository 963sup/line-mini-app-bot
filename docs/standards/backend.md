# 後端規範

後端使用 Go，負責可信身分、授權、業務規則、狀態轉換、資料一致性與外部整合。

## 套件責任

- `internal/business/<context>/domain`：實體、值、政策與純規則。
- `internal/business/<context>/application`：用例協調與使用端 port。
- `internal/business/<context>/adapter`：HTTP、LINE、Firestore 或其他技術轉譯。
- `internal/platform/identity` 提供可信使用者與驗證錯誤；其 `session` 套件管理建立、輪替、到期、續期與撤銷。Firestore 轉接器在交易內核對通道與 CSRF 後撤銷 Session。
- `internal/platform/line` 驗證 LINE token，`lineid` 統一群組及使用者 ID 格式。`internal/platform/google` 驗證服務呼叫的 Google token。資源所有權與授權由業務領域判定。
- `internal/transport/http/identity` 處理 Bearer、Cookie、Origin、CSRF 標頭與 HTTP 回應。`internal/bootstrap` 驗證部署設定並組裝驗證器與 Session 服務。
- `internal/platform`：不擁有業務語意的共用技術能力。
- `internal/transport`：共用傳輸入口。
- `internal/bootstrap`：依賴組裝與程序生命週期。

## Go 慣例

- 套件名稱使用 `lowercase`，例如 `navigation`、`richmenu`。
- 匯出識別字使用 `PascalCase`，例如 `AttendanceRichMenus`。
- 非匯出識別字使用 `camelCase`，例如 `loadAttendanceMenu`。
- 優先使用標準庫與小介面；介面由使用端定義。
- 傳遞 request `context.Context`，外部呼叫必須有期限並正確釋放資源。
- 錯誤保留操作上下文，但不得記錄 secret、access token、原始身分 token 或不必要個資。
- 公開資料格式使用契約產生型別；外部輸入在邊界驗證。
- 依賴使用建構函式明確注入，避免全域可變狀態與不必要的 DI framework。

## 寫入與外部副作用

交易 callback 可能重跑，因此不得直接在 callback 內傳送訊息或呼叫不可重複的外部動作。重要寫入需定義版本衝突、重送鍵、回應遺失與結果未知的處理。外部副作用要有明確的提交點與恢復策略。

## 設定

專案自定義環境變數使用 `SCREAMING_SNAKE_CASE`，例如 `LINE_MINI_APP_ENV`。環境變數只在 bootstrap 或專用設定載入器解析。必填值、型別、允許值與安全預設必須在載入邊界驗證；診斷只能列出設定名稱，不能輸出秘密值。

## 驗證

修改 Go 行為時，至少執行受影響套件測試，並依風險補上 `go test`、`go vet`、建置、race、漏洞或 Firestore emulator 驗證。格式化範圍只限本次擁有的檔案。
