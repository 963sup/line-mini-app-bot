# 驗證流程

驗證的目標是回答「這次變更的哪些行為已有證據」。命令成功、工具存在與環境配置存在都不是完整驗收。

## 選擇檢查

1. 依變更列出契約、行為、資料、安全、相容性與部署風險。
2. 執行最接近變更的單元或整合案例。
3. 執行必要的共用檢查。
4. 需要外部系統或裝置時，分開取得環境證據。

## 常用命令

```bash
git diff --check
go -C backend test ./...
go -C backend vet ./...
go -C backend build ./...
pnpm --filter @repo/miniapp check
pnpm --filter @repo/miniapp test
```

路由、靜態輸出或映像變更另執行對應 build。契約變更需驗證 OpenAPI、產生碼與兩端消費者。實際可執行入口以目前 `package.json`、Go module 與 CI workflow 為準。

## 後端架構測試

`go -C backend test ./...` 包含後端根套件的架構邊界與回歸測試。從儲存庫根目錄單獨重跑這兩項檢查時，可指定測試檔，讓掃描器在應用程式編譯失敗時仍能執行：

```bash
go -C backend test -count=1 architecture_test.go architecture_regression_test.go
```

## 整合證據

- Firestore 案例必須確認 emulator 已啟動、目標 test 實際執行且沒有 skip。
- LINE 驗證需區分 mock、測試 channel 與真實裝置。
- Cloud Run 驗證需記錄 revision、設定、流量與觀察窗口。
- 安全驗證需包含拒絕案例，不能只測 happy path。

## 報告

記錄命令、結果、失敗或 skip、環境與未驗證行為。不得把單一 probe 擴大宣稱為整個整合或部署通過。

## Firestore emulator 整合測試

在 Linux x86_64 的新環境，從專案根目錄執行 `scripts/test/firestore-emulator.sh`。腳本需要專案指定版本的 Go、Bash、curl、tar、sha256sum 與 flock。首次執行會下載並驗證固定版本的 Java 25 JRE 與 Firestore emulator JAR，之後重用本機快取。腳本啟動獨立模擬器，預設執行 Team 與群組整合測試，回傳測試退出碼並關閉模擬器。自動化可傳入其他 Go 測試，例如 `scripts/test/firestore-emulator.sh -run '^TestWikiMembershipAndRevisions$' ./internal/business/repository/adapter/http/wiki`。預設埠 18080 被占用時，可用 `FIRESTORE_TEST_PORT` 指定空閒埠。唯讀環境可將 `FIRESTORE_TOOL_CACHE` 指向可寫目錄。

等本機端點可達後，在另一個終端機指定 `FIRESTORE_EMULATOR_HOST`，用受影響的案例驗證。例如：

```bash
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 \
  go -C backend test \
  ./internal/business/repository/adapter/http/wiki \
  -run '^TestWikiMembershipAndRevisions$' -count=1 -v
```

## 專案建立重送與跨組織複製

先建置 MINI App 靜態頁面，啟動本機 Firestore emulator，並準備 Python Playwright 與 Chromium。在 backend 目錄執行：

```bash
MINIAPP_BROWSER_EXPORT="$(cd ../frontend/miniapp/out && pwd)" \
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 \
go test -count=1 -run '^TestProjectSettingsBrowserRoundTrip$' ./internal/business/project/adapter/firestore
```

瀏覽器送出建立個人專案後，測試等後端成功寫入才中斷回應。從畫面重試原請求後，核對開啟的專案 ID 與第一次寫入一致，並透過清單 API 確認個人範圍只有該筆專案。接著儲存欄位與檢視、複製到組織，再重新開啟兩份專案核對欄位名稱、選項及檢視引用。組織與專案資料經 Go HTTP 處理器讀寫 emulator；LINE 登入與 session 使用測試身分。

[瀏覽器驗收案例](../../backend/internal/business/project/adapter/firestore/browser_test.go)

## 文件候選覆核與寫入

- 用一份有文字層的文件和一張手機拍攝的工項表走同一覆核流程。候選可回查引文與來源位置，缺證時明確標示。
- 修正候選後只勾選一筆，應只建立這一筆，並顯示結果與物件連結。確認後撤銷寫入權限時，後端拒絕建立。
- 模擬寫入成功但回應遺失，使用相同建立識別碼重試，應回傳原結果且只保留一筆物件。

[物件契約](../../api/contract/object/objects.yaml)

## 定位權限與裝置驗收

- HTTPS／不安全來源、不支援 API、權限拒絕、OS 關閉定位、逾時、低精度、過舊／未來時間、缺欄／越界、離頁後 callback、跨群組越權。 受控測試使用 mock position；定位成功必須另有 LINE iOS/Android 與外部瀏覽器實機證據，不以桌面模擬取代。
