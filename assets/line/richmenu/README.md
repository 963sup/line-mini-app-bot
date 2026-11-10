# 基礎雙頁選單

## 自動同步指令

```sh
bash scripts/tooling/go-tool.sh run ./cmd/richmenu sync -images ../assets/line/richmenu -app-url https://miniapp.line.me/2012412007-0xCIzCmZ
bash scripts/tooling/go-tool.sh run ./cmd/richmenu sync -images ../assets/line/richmenu -app-url https://miniapp.line.me/2012412007-0xCIzCmZ -apply
```

預覽只驗證素材並顯示雜湊，不連 LINE。正式同步會比對 LINE 上以圖片和
完整設定 SHA-256 命名的選單，已有圖片就重用；兩頁圖片上傳完成後才更新
alias，最後設定出勤頁為預設並讀回確認。圖片不變且遠端狀態正確時不寫入。
新版 alias 與預設選單讀回確認後，將舊版個別使用者綁定批次遷移至對應新頁；
LINE 回報遷移成功才刪除舊版。只清理完整符合 managed-attendance-in/out 加
64 位雜湊命名的選單，且保護目前兩頁與任何仍被 alias 引用的選單。
清理失敗回傳錯誤，下次同步重試；不保留上一版作為固定備份。
部署應序列執行同步，避免不同版本同時更新。
alias 更新不是原子交易；中途失敗可重跑收斂，短暫期間可能混用新舊頁。
舊名稱不符合上述規則的選單需另行盤點，不會自動刪除。

容器包含 `/richmenu` 與 `/assets/line/richmenu`，可由獨立部署工作執行：
`/richmenu sync -images /assets/line/richmenu -app-url HTTPS_URL -apply`。
token 由部署工作的 Secret Manager 綁定注入，不放入映像或 build argument。

**目前 GitHub managed trigger 尚未接入此指令。** 只換圖片並推送會將圖片
打包進映像，但不會自動修改 LINE。需要在實際部署流程加入部署後的獨立工作，
設定 token 與上述 command；本次未修改外部 trigger、建立 Cloud Run Job 或發布 LINE。
不在 HTTP server 啟動時同步，以免多實例與舊 revision 冷啟動覆蓋新版設定。

兩張 JPEG 均為 2500 × 843，三個點擊區域與圖片等寬對應。
attendance-in：本人出勤紀錄／切換至功能頁／MINI App。
attendance-out：功能說明／切換至出勤頁／MINI App。

先預覽，再依序發布兩張圖片；第二張發布時設定預設選單。

```sh
bash scripts/tooling/go-tool.sh run ./cmd/richmenu -app-url https://miniapp.line.me/2012412007-0xCIzCmZ -attendance-preview
bash scripts/tooling/go-tool.sh run ./cmd/richmenu -app-url https://miniapp.line.me/2012412007-0xCIzCmZ -alias attendance-in -image ../assets/line/richmenu/attendance-in.jpg -apply
bash scripts/tooling/go-tool.sh run ./cmd/richmenu -app-url https://miniapp.line.me/2012412007-0xCIzCmZ -alias attendance-out -image ../assets/line/richmenu/attendance-out.jpg -apply -set-default
```

發布需要環境變數 LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN。失敗後使用印出的
`-menu-id` 重試；既有 alias 若指向其他選單，會停止，避免覆寫既有設定。
切換由 LINE 的 richmenuswitch action 執行，不需伺服器切換使用者選單。
查詢按鈕送出 attendance-history postback，由伺服器依簽章事件的本人 ID
列出有紀錄的群組，再以「出勤紀錄 編號」取得該群最近五筆。

契約來源：Context7 /line/line-openapi 與 /line/line-bot-sdk-go/v8.23.0；
LINE 文件來源 main 090500a12f9777995006d9fb5cafab8d83a792b2。
