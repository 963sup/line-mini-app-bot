# 架構測試（Architecture Testing）

把依賴與邊界約束轉為可自動檢查的規則。

## 設計時要確認

對 Go import 圖檢查 domain 是否依賴 SDK 和 Adapter，對埠另做契約測試；例外須具名並有理由。

## 示例

一個測試抓 domain 引入 Firestore，另一個以真正儲存驗證版本衝突。

## 用同一組案例驗證兩種 Adapter

假設出站 Port 承諾依預期版本更新：先讓兩個寫入者讀到相同版本，再各自提交不同內容。驗收要看最終狀態、成功次數、衝突語意與失敗方有無部分寫入。把同一組案例交給記憶體替身和儲存 Adapter；替身可重現用例分支，儲存實作另須在其支援的交易環境驗證。若替身總是成功，通過的用例測試無法支持併發保證。

| 要判斷什麼 | 選用證據 | 仍需另外確認 |
| --- | --- | --- |
| 規則與時間邊界 | 固定時間，測到期前、當下與之後的領域結果 | 資料庫是否實施同一衝突條件 |
| 用例與外部能力是否分離 | 替換出站 Port，注入拒絕、失敗及未知結果 | 真實 Adapter 的錯誤映射、重試與時限 |
| 公開介面是否相容 | 對照既有請求、狀態碼、錯誤與產生碼 | 使用者是否有權操作實際資源 |
| 原始碼依賴方向 | [Go 邊界測試](../../../backend/internal/architecture/boundaries_test.go) | 規則是否真的由正確的上下文負責 |

現有出勤用例與儲存測試可從[補正審核案例](design-guide.md#把概念用到現行程式補正審核)追查。上述版本衝突案例是設計驗收範例，需讀對應測試的斷言才能列為已覆蓋。

## 延伸閱讀

[dependency-rule](dependency-rule.md) · [domain-testing](domain-testing.md) · [driven-adapter](driven-adapter.md) · [DDD 設計索引](design-guide.md)
