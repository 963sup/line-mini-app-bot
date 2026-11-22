# 查詢、分頁與索引

從實際操作需要的範圍、篩選與排序設計查詢，再維護索引；不能先建立通用查詢介面再讓前端拼出授權。本頁為現行查詢入口與維護指引。

## 查詢契約

每個列表寫明 Actor、Repository 或其他作用域、篩選、排序方向、固定的同值排序鍵、頁大小與游標失效條件。列表結果不會授予後續寫入資格；詳情或異動仍依其用例重新核對。

[Repository store](../../../backend/internal/business/repository/adapter/firestore/store.go)以文件 ID 分頁使用者候選；[Object store](../../../backend/internal/business/object/adapter/firestore/store.go)以文件 ID 排序，可附父項篩選；[Wiki store](../../../backend/internal/business/repository/adapter/firestore/wiki.go)的頁面列表與修訂列表有不同的排序與游標。沿各自實作核對，不能用一種游標套用所有查詢。

## 游標與可見性

游標須符合對應操作的身分、作用域、篩選及排序約定；解析成功不代表有讀權。對外格式依功能契約，不直接讓客戶端決定 Firestore 路徑。

分頁過程資料可能變動，除非有具體快照機制，不宣稱跨頁一致快照。取得候選後再核驗資格可能形成空頁；仍有下一頁時，UI 應依契約允許繼續。不能把單頁長度當成全庫總量。

## 索引與部署

索引與欄位設定集中於 [firestore.indexes.json](../../../firestore.indexes.json)。目前包含出勤複合索引及 Session 的 TTL／索引設定；檔案存在不代表目標資料庫已套用或建置完成。

新增查詢先核對篩選／排序及官方[索引說明](https://firebase.google.com/docs/firestore/query-data/index-overview)，再決定必要宣告。索引建立、資料回填、讀端切換與回復條件分開記錄，正式套用依[發布驗收](../../guides/release/release-acceptance.md)執行。

## 成本與效能

觀測每頁讀取量、核驗外部服務次數、延遲與熱點，設定有依據的頁大小及界線。避免無界掃描、將整個集合載入記憶體再篩選，或為查詢方便複製無明確更新責任的資料。需要投影時，定義來源、更新方式及可接受延遲。

## 驗收

檢查空結果、同值排序、下一頁、非法游標、跨本人／跨庫／跨篩選重用、資料增刪及權限失效。emulator 上查詢成功不能證明正式複合索引已存在；另確認目標環境、索引狀態及代表負載。未做正式觀測時列為未驗證。

[資料配置](data-model.md) · [交易與一致性](consistency.md) · [資料生命週期](lifecycle.md) · [儲存索引](README.md)
