# Cloud Storage 檔案儲存簡要規格

**狀態：工程附件的候選方案。** 保存 PDF、工程圖、照片與匯出檔案；Firestore 保存附件識別、擁有者、作用域及檔案版本關聯。沿用[工程平台提案](../../proposals/line-engineering/README.md)的分工，不宣稱上傳功能已實作。

## 最小流程

1. Go 驗證 Actor 對目標 Repository／資源的上傳權，建立 pending 附件及限定的 object key。
2. 小檔可經 Go 上傳；大檔依需求使用短效 Signed URL 或受控續傳。先限制宣告的類型、大小及目的地，再核驗實際物件。
3. 上傳完成後，後端核對 bucket、object、generation、實際大小及允許的內容，才把附件標為可用並建立業務關聯。需安全掃描的內容先隔離。
4. 下載時重新核對讀權，再由 Go 傳送或產生短效下載 URL；背景處理綁定確切檔案版本。

## 資料與授權邊界

Bucket 預設不公開，IAM 及簽署身分採最小權限。Signed URL 在有效期間具有持有者存取能力，不放日誌或當作永久附件地址；不能宣稱撤回業務權限會立即撤銷已簽出的 URL。

Firestore 與 Storage 沒有共同業務交易，以 pending／ready／failed 等明確生命週期協調（正式欄位及狀態於實作時確認）。登記成功但上傳失敗、檔案存在但 metadata 未提交，均須可對帳及清理；不得為覆寫原檔而沿用不受控的路徑。

## 導入與保存

指定 bucket、位置、CORS、大小／格式限制、URL 期限、檔案保留與清除責任。Firebase Storage 以 Cloud Storage 為基礎，選定一套存取及授權方式；既有 Go Session 不因此自動具備 Firebase 客戶端權限，也不強制另建一套登入。

物件版本、刪除及恢復政策依[資料生命週期](../firestore/lifecycle.md)決定；Maps 或其他第三方資料的保存另遵循來源條款。

## 驗收

檢查跨庫越權、過期 URL、超限／錯誤內容、重送完成通知、同名覆寫、舊 generation、孤兒檔案與下載撤權。驗證上傳後 metadata 和物件版本一致；雲端 IAM、CORS 及生命週期需在目標環境另行驗收。

## 來源與相關規格

[官方 Signed URL 說明](https://docs.cloud.google.com/storage/docs/access-control/signed-urls)，查證範圍見[Google 整合總覽](../google-cloud/README.md#sources)。

[Eventarc](../google-cloud/eventarc.md) · [Document AI](../google-cloud/document-ai.md) · [儲存索引](../../architecture/data.md)
