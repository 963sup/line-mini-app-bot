# Summary 既有入口契約

本文件描述既有 LINE／內部工作入口，不加入公開 OpenAPI、不產生瀏覽器 client。
完整產品規則與資料保存方式見 [群聊摘要](../../../docs/features/summary.md)。

| 能力 | 入口與可信來源 | 輸入、規則與結果 | 後端 |
| --- | --- | --- | --- |
| MessageCollection | 共用 LINE webhook 驗證簽章後的群組文字／編輯／收回事件 | 一般文字靜默收集；指令排除；編輯與收回沿用既有版本失效規則 | business/summary/messagecollection、其 line/；共用 summary/firestore |
| Summarization | 同一 webhook 的群組「摘要」指令 | 有效群組、使用者、event ID、reply token 及既有時效檢查；event ID 防重，建立工作並排程；完成工作不重排 | business/summary/summarization/line |
| Summarization worker | POST /internal/summary；Google 工作身分 Bearer token 驗證 | JSON group（C 加 32 位小寫十六進位）、id（64 位小寫十六進位）；2048 bytes 上限、拒絕未知欄位與多份 JSON | business/summary/summarization/http |

worker 保持現有回應：204 無內容；401 unauthorized；400 invalid job；503 summary temporarily unavailable。
佇列傳送小寫 group、id；handler 沿用 Go JSON decoder 的大小寫匹配，不藉重構改變接受格式。
工作沿用租約、版本重查、結果保存與穩定 LINE retry key；模型或發送失敗交由既有工作重試。
共用 LINE 組合入口在 business/summary/line，依序交給摘要指令與訊息收集；不讓收集模組依賴摘要生成。

沒有既有 MINI App 畫面或公開 Summary HTTP API；不得把此內部工作端點當成瀏覽器 API。
