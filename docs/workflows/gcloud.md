# Google Cloud CLI 安裝與登入

當使用者要求執行需要 Google Cloud CLI 或本機 Google Cloud 憑證的工作時，Codex 依此流程檢查工具、安裝缺少的 CLI、驗證目標並執行已授權的工作。需要 Google 帳號授權時，由使用者在瀏覽器完成。

## 安裝

從儲存庫根目錄執行 [安裝腳本](../../scripts/setup/install_gcloud.sh)。腳本會重用 PATH 上的 `gcloud`，或從 Google 官方下載 Linux x86-64 或 ARM64 SDK 到可寫入的目錄。預設安裝在 `~/.local/share/google-cloud-sdk`。可傳入第一個參數指定安裝根目錄。腳本需要 Bash、curl 與 tar。

```bash
bash scripts/setup/install_gcloud.sh
```

新安裝時，依腳本輸出的指令更新目前 shell 的 PATH，或以安裝目錄下的 `bin/gcloud` 絕對路徑執行。每次工具呼叫若會啟動新 shell，使用絕對路徑即可。其他平台使用 [Google 官方安裝方式](https://docs.cloud.google.com/sdk/docs/install-sdk)。

若沙箱限制家目錄寫入，將安裝根目錄指定為儲存庫外、可寫入且只有你可存取的位置，並將 `CLOUDSDK_CONFIG` 指向該位置下的私有設定目錄。登入與後續每次工具呼叫都使用同一設定路徑。臨時環境結束後，憑證是否可重用取決於該儲存位置是否保留。

## 登入

1. 優先使用同一執行環境中已具備的憑證。CLI 工作可用下列指令檢查目前選取的帳號。

   ```bash
   gcloud auth list --filter=status:ACTIVE --format='value(account)'
   ```

2. 需要 CLI 帳號登入時，Codex 在使用者可輸入的互動終端啟動下列指令，顯示 Google 授權網址，並保留原程序等待輸入。

   ```bash
   gcloud auth login --no-launch-browser
   ```

3. 使用者開啟網址並授權後，把一次性授權碼直接輸入原終端。授權碼、存取權杖與憑證檔內容應留在登入環境，避免貼到聊天或寫入日誌。若 Codex 的終端無法由使用者操作，由使用者在同一執行環境及同一帳號下的終端執行完整登入指令。
4. 登入成功後，Codex 重新確認帳號，並對已授權的工作使用明確的專案 ID。帳號已選取只表示 CLI 有設定帳號，實際權限以目標操作回應為準。

Go／Firestore 等本機程式若使用 Application Default Credentials（ADC），登入步驟改用 `gcloud auth application-default login --no-launch-browser`。可執行 `gcloud auth application-default print-access-token >/dev/null` 檢查 ADC 能否取得權杖，再驗證目標資源的權限。

官方參考 [CLI 登入](https://docs.cloud.google.com/sdk/gcloud/reference/auth/login)與 [ADC 登入](https://docs.cloud.google.com/sdk/gcloud/reference/auth/application-default/login)。
