# 外部系統

本頁是整合導航，不複製供應商完整文件。行為與版本必須以 manifest、目前實作及對應版本的官方文件核對。

| 系統                                      | 專案用途                               | 主要邊界                                    |
| ----------------------------------------- | -------------------------------------- | ------------------------------------------- |
| LINE Messaging API                        | Webhook、Reply／Push、群組與 Rich Menu | 簽章、token、重送、reply token 期限         |
| LINE MINI App／LIFF                       | 使用者登入與應用入口                   | 後端 token 驗證、channel 對應、真實裝置     |
| Cloud Run                                 | Go service runtime                     | revision、service account、traffic、timeout |
| Cloud Build／Artifact Registry            | 映像建置與保存                         | immutable image、build identity、供應鏈     |
| Firestore                                 | 業務資料與 Session                     | transaction、index、TTL、IAM                |
| Cloud Storage                             | 檔案與大型物件                         | object name、IAM、版本、保存                |
| Cloud Tasks                               | 背景摘要等工作                         | OIDC、audience、重試、operation ID          |
| Secret Manager                            | runtime 機密                           | 明確版本、最小存取、輪替                    |
| OpenAI／Gemini                            | 摘要、Agent、文件候選                  | 資料最小化、模型輸出驗證、timeout           |
| CWA                                       | 天氣資料                               | API key、供應端錯誤與資料時效               |
| OpenTelemetry／Google Cloud Observability | log、metric、trace                     | 敏感資料、label cardinality、版本           |

## 查證原則

外部能力變更時，先確認專案實際版本，再查官方文件；來源不可用時列明限制。官方文件說明能力，不代表此專案已設定、授權、部署或通過真實環境驗收。
