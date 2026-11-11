# 群組請假登記

## 最小流程

在 LINE 群組輸入「請假」，Bot 回覆「填寫請假」按鈕。連結綁定原群組，有效30分鐘；成員開啟 MINI App、完成 LINE 本人驗證，填寫假別、臺灣時間的起訖與理由後送出。請假不需要定位。

假別為事假、病假、特休、其他。理由必填，最多500字；結束須晚於開始，單筆期間最多366天，開始年份限制2000至2100。允許登記過去時段，不計算應到日或可用假額。

**第一版僅登記，不代表核准。** 沒有審核、扣假、撤回、通知或自動發布群組結果；請另行聯絡主管確認。請假不改寫打卡、工時或更正資料，不以假單推定出勤。理由及假別不回傳群聊。

本人可在 `/miniapp/leave` 或首頁「本人請假紀錄」查看跨群組最近20筆登記；畫面顯示群組ID供區分。沒有主管查看全群假單的權限或入口。新增申請須從原群組取得連結，不能自行指定群組或他人 user ID。

## API 與儲存

- `POST /api/v1/attendance/leaves`：launch、requestKey（UUID）、kind、start、end、reason。時間為帶時區 RFC3339。回傳已保存的 LeaveRecord，status 固定為 `registered`，不是 approved。
- `GET /api/v1/attendance/leaves`：登入本人最近20筆登記，按建立時間遞減，不代表完整歷史列表。
- 沿用既有 Cookie Session／Bearer 認證、CSRF 與 OpenAPI 請求驗證。後端以 LINE API 確認原群組成員，只有200才允許送出；非成員拒絕，API故障不寫入。
- 同一本人的 requestKey 在相同內容重試時回原登記；不同內容回409。前端在結果未知時鎖住草稿並重試原內容；成功後才可明確建立另一筆。
- 回應逾時不代表失敗，先重新載入本人紀錄。頁面重開、連結過期後，應先確認最近紀錄再重新填寫；第一版不保存跨頁草稿，也不提供語意重複／時段重疊的自動拒絕。
- Firestore：`channels/{channel}/leaveLaunches/{tokenHash}` 保存群組與到期；`channels/{channel}/users/{userHash}/leaveRequests/{requestKeyHash}` 保存伺服器以 `google/uuid` 產生的 UUID v4 申請編號、群組、token雜湊、假別、UTC起訖、理由及建立時間。原始 token 不寫入假單文件；交易再次確認連結有效，重試不建立第二筆。
- 最近紀錄僅在本人子集合以單欄 createdAt 排序並限制20筆，不需要新增複合索引、回填或新環境變數。保存期與清除政策尚未設定；不擅自替請假理由或去重紀錄加TTL。

MINI App 與 Messaging API 仍需同一 Provider，且 Bot 在群組中。沿用現有 LINE_MINI_APP_LIFF_ID、身分／Session 與出勤資料庫設定。主頁的 `?leave=` 及 LIFF `liff.state` 會導向請假頁，打卡連結維持原流程。

## 查證與驗證邊界

本次查核 LINE 官方來源庫 default branch main，revision `dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54`：
[群組訊息與群組來源](https://developers.line.biz/en/docs/messaging-api/group-chats/)、[伺服器驗證本人](https://developers.line.biz/en/docs/liff/using-user-profile/)、[LIFF連結與附加參數](https://developers.line.biz/en/docs/liff/opening-liff-app/)。

交易讀取先於寫入與重試語意使用 Context7 `/websites/cloud_google_firestore_native` 查核[官方交易文件](https://docs.cloud.google.com/firestore/native/docs/manage-data/transactions)，對照專案Firestore Go v1.29.2；前端參照 Context7 `/react/react/v19.3.1` 與已安裝 Next.js 16.4.5 Pages Router 指引。未新增依賴。

本機檢查不能代替 LINE iOS／Android、真實群組成員 API 或正式 Firestore 寫入驗收。Git 推送可能觸發既有CI/CD，不等於部署已成功；本次不手動操作雲端設定。

[功能索引](README.md) · [出勤功能](attendance.md) · [文件總索引](../README.md)

## 本次本機驗證

已通過 fmt-check、lint（0 issues）、mise check（mod verify／vet／既有 race suite）、mise build、前後端 pnpm build、TypeScript check、12個既有前端測試、API生成同步／相容性（含 origin/main 歷史比較）、設定檢查與 git diff --check。govulncheck 沒有呼叫路徑中的漏洞；相依模組仍報3個未被呼叫的漏洞。

啟動未設定雲端的本機服務，確認 `/miniapp/leave`、主頁 `?leave=`、`liff.state` 請假連結均回200並呈現請假頁，原打卡連結仍呈現打卡頁；未設定時請假 API 回503。Graphify 已更新。

沿用既有測試，未新增測試工程或案例。Firestore emulator 未設定，既有整合測試明確 skip；新增請假交易與真實 LINE 成員檢查仍須外部驗收，不把既有測試通過視為新流程整合已通過。

新申請的 API `id` 為 UUID v4；文件路徑保留 requestKeyHash 以維持原子防重複。舊資料沒有 `id` 時仍回傳原文件編號，不改寫歷史資料。UUID 產生失敗時交易不寫入。
