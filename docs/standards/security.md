# 安全規範

安全判斷由後端在信任邊界執行，預設拒絕未驗證或未授權操作。

## 信任邊界

| 入口           | 必要驗證                                              |
| -------------- | ----------------------------------------------------- |
| LINE Webhook   | 以對應 channel secret 驗證原始 body 簽章              |
| MINI App 登入  | 後端驗證 LINE token、預期 channel、有效期與簽發條件   |
| 瀏覽器業務請求 | Cookie Session、Origin／跨站保護與 CSRF；再做資源授權 |
| 內部工作端點   | Google ID token、正確 audience 與允許的服務帳戶       |
| 公開診斷       | 保持無狀態，不授予身分或業務權限                      |

網址參數、postback data、前端角色、LINE profile 或畫面上是否顯示按鈕都不是授權證據。Session 證明操作者，不取代群組、成員或資源所有權檢查。

## Session 與瀏覽器

Session cookie 必須使用 Secure、HttpOnly、Path=/、SameSite 與 host-only 限制。CSRF token 不寫入持久儲存。登出、到期與身分切換後必須清除私有畫面並拒收舊請求結果；登出不等同取消已被後端接受的交易。

## 秘密與個資

- 秘密透過 Secret Manager 或受控執行環境注入，不進入映像、儲存庫、URL 或一般 log。
- 使用明確 secret version，保留安全回退版本並驗證新 revision Ready 後才切換。
- 記錄使用 bounded 類別、correlation ID 與必要資源識別；排除 token、cookie、原始 payload 與不必要個資。
- 憑證可用不代表具有部署、IAM 或資料變更授權。

## 輸入與外部呼叫

所有外部輸入都需限制大小、格式、路徑與資源範圍。外部 HTTP 呼叫設定 timeout、限制 redirect，並將供應端錯誤轉成安全且穩定的內部分類。

## 驗證

安全相關變更需涵蓋無效憑證、過期、錯誤 audience、CSRF、越權、跨租戶、重送及敏感資料不外洩。mock 或編譯不能代替真實 IAM、LINE channel 或瀏覽器 Cookie 驗收。
