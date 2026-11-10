# Monorepo 整合、發布與回復

本頁是可執行步驟與驗收表，不是正式環境已通過的聲明。Cloud Run 是唯一部署目標；不使用 Firestore 模擬器。本機 mock、容器啟動、revision Ready 都不能替代 LINE iOS／Android 與真實資料流程。

## 發布證據與前置條件

1. 選定 project／region／service，取得有效 gcloud 身分；每個命令明確帶 project／region，不依賴目前可能錯誤的預設值。不要在對話提供 token、私鑰或 Secret payload。
2. 記錄候選來源 commit、工作目錄差異與映像不可變 digest。未提交來源建置的映像不能只用 HEAD SHA 識別。先跑既有 pnpm check／test／build；測試環境不要設定 FIRESTORE_EMULATOR_HOST。
3. 暫停／協調其他部署，核對 managed trigger：main push 會自動部署；它不等待本驗收流程。不要與手動部署同時執行。
4. 取得發布前 snapshot，記錄所有 serving revisions 與流量分配、image digest、service account、Secret 數字版本。核對環境變數實際值；snapshot 刻意只列 env 名稱，不能單靠它判定 Channel／LIFF／Firestore 設定正確。
5. 複查 Cloud Run IAM、Secret／Firestore／Cloud Tasks 所需權限與 LIFF Endpoint、openID scope、Messaging webhook。目標專案出現資料不代表目前候選版本已驗收。若需測試寫入，使用明確的測試群組／使用者；不刪改正式出勤紀錄。

```bash
# 讀取服務與涉及的 revision metadata；不讀取 Secret payload。
bash scripts/tool-exec.sh python3 scripts/cloud-run-snapshot.py \
  --project=YOUR_PROJECT --region=asia-east1 --service=line-mini-app-bot \
  > /tmp/release-before.json

# HTTP 無登入 smoke：不送 token、不建立出勤或申請。
python3 scripts/release-smoke.py --base-url=https://YOUR_SERVICE_URL \
  > /tmp/release-http-before.json
```

snapshot 需先成功才進行發布；輸出不含 env 明文值，但仍應以內部交付資料管理。HTTP smoke 會檢查 Next 頁面、CSP、資源、未認證拒絕與不存在路徑。它不初始化真實 LIFF，不驗證 GPS、資料庫或完整業務流程。

本機未設定資料庫的容器可加 `--allow-disabled-attendance`，明確允許出勤端點回 503；此選項只接受 localhost，不得用於目標環境驗收。它不代表出勤可用。

## 候選建置與無流量發布

依現有部署路徑準備候選映像，先完成來源上傳清單與 frozen-lockfile 建置。部署憑證和 LINE 憑證只在執行期使用，不能放進 source、build args 或 image。不要為測試改動正式 Secret、IAM 或資料庫模式。

在既有服務、既有完整 runtime 設定已核對的前提下，用不可變 digest 建立候選 revision：

```bash
# 會變更雲端；不是 dry run。替換所有 YOUR_*，取得本次發布授權後執行。
bash scripts/tool-exec.sh gcloud run deploy line-mini-app-bot \
  --project=YOUR_PROJECT --region=asia-east1 \
  --image=YOUR_REGISTRY_IMAGE@sha256:YOUR_DIGEST \
  --revision-suffix=YOUR_UNIQUE_SUFFIX --tag=acceptance --no-traffic
```

這條 image-only 命令沿用服務設定，不適用尚未完成設定的首次部署。不要加入 `--set-env-vars`／`--set-secrets` 重建未核對的集合。現有 cloudbuild.yaml 的 deploy step 會直接更新服務，並非無流量驗收流程；若採該檔發布，先調整策略再送出，不能假設它有上述保護。

候選 tag URL 仍有原服務 IAM／ingress 限制。LIFF 必須在已設定 Endpoint URL 或其下層初始化；不能把正式 LIFF 任意指向 tag hostname。真實裝置驗收使用已核准且已配置 LIFF 的測試環境，或依既定變更時段切換正式流量驗收；不得以 mock SDK 當作真實 LINE。此取捨須由環境擁有者確定，避免為測試任意改正式 Endpoint。

## 真實裝置驗收表

每項在 iOS LINE、Android LINE 分別記錄裝置／OS／LINE 版本、時間、測試者、revision、結果與必要的匿名化證據；另測 Safari／Chrome 外部登入。不把 ID token、原始定位、使用者 profile 或完整出勤內容放入一般日誌。

| 流程 | 驗收條件 | iOS | Android |
|---|---|---|---|
| 群組打卡連結 | 經真實訊息啟動，launch／liff.state 保留；init 完成前不修改 SDK URL | 待測 | 待測 |
| 身分 | 真實 openid token 經後端驗證；外部登入返回原頁；錯 Channel／過期不取得資料 | 待測 | 待測 |
| 定位 | 允許、拒絕、撤銷後重試、逾時；離頁／背景切換不留下 watcher 或晚到打卡 | 待測 | 待測 |
| 打卡 | 測試群組完成上班／休息／下班，資料與本人紀錄一致；成功才關窗 | 待測 | 待測 |
| 弱網重試 | 提交回應遺失後手動重試，同一 requestId 不重複寫入；不自動重送 | 待測 | 待測 |
| 月紀錄 | 群組／月份切換、前後分頁、完整月總計、跨月班次，無舊資料覆蓋 | 待測 | 待測 |
| 更正 | 同草稿重試僅一單，內容變更換碼；未核准工時不變 | 待測 | 待測 |
| 審核 | 另一授權帳號讀取並核准；本人／未授權被拒絕；更換編號清除舊內容 | 待測 | 待測 |
| 登出／返回 | 清空私人資料，BFCache 返回重新驗證，不顯示上一身分資料 | 待測 | 待測 |
| LINE Messaging | 一對一／群組命令、Webhook 驗簽與 redelivery 不改既有行為 | 待測 | 待測 |

真實寫入後記錄測試資料標識與處置方式。不得以直接刪除正式資料作清理；出勤更正採現有審核契約。

## 切換、觀察、回復

先記錄具名舊 revision（或原流量分配圖），候選通過後以具名 revision 切換。不能使用 `LATEST` 作回復識別。

```bash
# 會變更正式流量；完成候選驗收及發布授權後執行。
bash scripts/tool-exec.sh gcloud run services update-traffic line-mini-app-bot \
  --project=YOUR_PROJECT --region=asia-east1 --to-revisions=YOUR_ACCEPTED_REVISION=100

# 讀取狀態，執行 HTTP smoke、裝置工作流程並觀察錯誤／延遲／未授權結果。
bash scripts/tool-exec.sh python3 scripts/cloud-run-snapshot.py \
  --project=YOUR_PROJECT --region=asia-east1 --service=line-mini-app-bot > /tmp/release-after.json

# 回復：適用原本 100% 在單一 revision 的基線；若原本分流，還原原分配圖。
bash scripts/tool-exec.sh gcloud run services update-traffic line-mini-app-bot \
  --project=YOUR_PROJECT --region=asia-east1 --to-revisions=YOUR_PREVIOUS_ACCEPTED_REVISION=100
```

出現登入失敗、CSP／資源錯誤、重複寫入、查詢結果不一致或明顯錯誤率／延遲惡化即停止擴大發布，依既定門檻回復。回復後重跑 snapshot、HTTP、真實登入／查詢，確認流量與 image digest。保留原映像、數字 Secret 版本與舊 revision。

流量回復不撤回已寫入的 Firestore 資料、不撤銷 LINE 訊息，也不復原外部 Channel 設定。此次遷移未執行 schema 回填或最佳化讀取切換；若另行切換，須另備相容／回復計畫。相容 JS alias 是空操作，舊瀏覽器頁面需重新整理。

## 查證與限制

gcloud 592.0.0 的 run deploy／services describe／revisions describe／services update-traffic leaf help 已查。Google Developer Knowledge MCP 在本會話未提供，改查官方 Cloud Run [發布與回復](https://docs.cloud.google.com/run/docs/rollouts-rollbacks-traffic-migration)、[部署](https://docs.cloud.google.com/run/docs/deploying)；本機 cloud-run skill 僅用於 CLI 發現。LINE 官方來源 main 為 `dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54`，參照 [LIFF 初始化與 URL 注意事項](https://developers.line.biz/en/docs/liff/developing-liff-apps/#initializing-liff-app)。這些來源與 CLI help 不證明目標 IAM、裝置或部署驗收已通過。
