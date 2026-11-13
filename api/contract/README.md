# 契約能力索引

本目錄以能力歸屬組織契約。公開 HTTP、事件與內部工作入口有不同形式，不能因歸類而新增公開 API。功能授權與業務規則仍由對應功能文件定義。

| 能力／責任 | 來源 | 範圍 |
| --- | --- | --- |
| 公開 HTTP 組合入口 | [openapi.yaml](openapi.yaml) | API metadata、既有路徑與 components 的 `$ref` 索引 |
| 共用傳輸契約 | [common/transport.yaml](common/transport.yaml) | APIError、Bearer 與 Cookie security schemes；不承載業務模型 |
| Identity | [identity/identity.yaml](identity/identity.yaml) | MINI App 公開設定、本人、Session 建立／續期／撤銷及其 DTO |
| Repository | [repository/repositories.yaml](repository/repositories.yaml) | 容器列表、詳情、分頁及容器 DTO；不提供手動建立或管理 API |
| Space | [space/spaces.yaml](space/spaces.yaml) | 獨立群組空間列表／詳情與當下成員核驗；不依Repository服務授權 |
| Project | [project/space-projects.yaml](project/space-projects.yaml) | Space擁有的專案與items，來源唯讀；[Project規格](../../docs/features/projects.md) |
| Project legacy | [project/projects.yaml](project/projects.yaml)、[items.yaml](project/items.yaml) | 原Repository-scoped路由與DTO相容轉接；同一用例，不構成新擁有權 |
| Repository / Labels | [repository/labels.yaml](repository/labels.yaml) | 標籤目錄列出與建立，規格見 [標籤](../../docs/features/labels.md) |
| Issue | [issue/issues.yaml](issue/issues.yaml) | 空間事項列表、詳情、建立及建立者修改；操作規格見 [事項](../../docs/features/issues.md) |
| Discussion | [discussion/discussions.yaml](discussion/discussions.yaml)、[comments.yaml](discussion/comments.yaml) | 主題讀寫及單層留言；作者編輯，規格見 [討論](../../docs/features/discussions.md) |
| Attendance / Punch | [attendance/punch.yaml](attendance/punch.yaml) | launch、打卡與 preflight |
| Attendance / Records | [attendance/records.yaml](attendance/records.yaml) | 群組、出勤紀錄、月摘要及分頁 |
| Attendance / Worktime | [attendance/worktime.yaml](attendance/worktime.yaml) | 工時計算與時間區間型別，沒有獨立 HTTP 操作 |
| Attendance / Corrections | [attendance/corrections.yaml](attendance/corrections.yaml) | 補卡提交、列表、狀態、本人撤回；申請快照、狀態結果與共用補卡錯誤回應 |
| Attendance / Approvals | [attendance/approvals.yaml](attendance/approvals.yaml) | 審核明細、核准、駁回與審核群組；引用 Corrections 的申請資料 |
| Summary | [summary/README.md](summary/README.md) | MessageCollection、Summarization 與內部 worker，未加入公開 OpenAPI |
| Weather | [weather/README.md](weather/README.md) | Forecast 的既有 LINE 事件入口，未加入公開 OpenAPI |

## 巢狀指引

本目錄遵循 [contract 指引](AGENTS.md)，繼承 [API 指引](../AGENTS.md)；各能力補充如下：

| 範圍 | 指引 | 狀態 |
| --- | --- | --- |
| Common | [AGENTS.md](common/AGENTS.md) | 共用傳輸契約 |
| Identity | [AGENTS.md](identity/AGENTS.md) | 既有公開契約 |
| Repository | [AGENTS.md](repository/AGENTS.md) | Bot 管理的群組空間 |
| Issue | [AGENTS.md](issue/AGENTS.md) | 基礎事項 API；types／comments 仍為雛形 |
| Discussion | [AGENTS.md](discussion/AGENTS.md) | 主題／單層留言已接入；categories 仍為雛形 |
| Project | [AGENTS.md](project/AGENTS.md) | Project／items 已接入；fields／views 為雛形 |
| Attendance | [AGENTS.md](attendance/AGENTS.md) | 既有出勤契約 |
| Summary | [AGENTS.md](summary/AGENTS.md) | 事件與內部工作契約 |
| Weather | [AGENTS.md](weather/AGENTS.md) | 既有事件入口契約 |

## 歸屬與依賴

命名遵循 [業務實體、集合與 API 命名](../../docs/development/naming.md#業務實體集合與-api-命名)：單一實體型別用單數，資源集合及 REST 集合路徑用複數。既有單例及公開契約相容例外按該規範處理；能力名稱不以英文單複數機械判定。

- 根入口只組合來源，不內嵌能力的 path、schema 或 security scheme 定義。路徑與原有 component 名稱仍保留在入口索引，維持公開介面。
- 能力來源不得引用 `openapi.yaml`；共用錯誤與認證定義由 common 擁有。security requirement 名稱由根入口的 securitySchemes 組合解析，能力檔不是獨立完整 OpenAPI 文件。
- 能力擁有自己的操作、輸入及結果；真正跨能力共用的業務資料引用其擁有者，不能全部搬到 common。
- 出勤依賴方向：Approvals → Corrections → Records → Worktime；Approvals 亦直接引用 Records，Punch／Leave 引用 common。Corrections 不依賴 Approvals。
- `CorrectionReview` 名稱保持相容，但內容是列表與審核共用的補卡申請快照，因此由 Corrections 擁有；`CorrectionDecisionResult` 供撤回與審核使用，同樣由申請生命週期擁有者保存。
- 單一能力檔可包含相關 paths 與 schemas，不要求每個 endpoint 一個檔案，也不為沒有公開入口的能力建立空 HTTP 契約。

## 修改與驗證

### 最小契約雛形

Repository milestones／statistics 與 Project／items 已依 [基礎追蹤規格](../../docs/features/planning.md) 接入。以下檔案目前以 `x-status: scaffold` 標記，只有空的 `paths` 與 `components.schemas`，尚未定義端點或資料欄位，也不由公開／內部組合入口引用。能力清單登記其擁有者，不表示功能已實作。

- Issue：[types.yaml](issue/types.yaml)（IssueType）、[comments.yaml](issue/comments.yaml)。
- Discussion：[categories.yaml](discussion/categories.yaml)。
- Project：[fields.yaml](project/fields.yaml)、[views.yaml](project/views.yaml)；projects／items 已接入。

以上是能力契約片段，不是可獨立啟動或呼叫的完整 OpenAPI 文件。實作時先補功能操作規格，再填入對應 paths／schemas、接入組合入口、更新能力清單與生成碼，並驗證相容性；既有 Repository 仍由 Bot 建立。

先改來源契約，再依 [API 工作流程](../README.md) 生成共用 Go／TypeScript 產物。Redocly 從根入口 bundle，不需每個能力各建 client 或生成器；相容性 baseline 不隨搬移刷新。

純歸屬重組先比較搬移前後 `redocly bundle --dereferenced` 的完整 JSON 結構，再執行 `pnpm api:check` 與 `pnpm api:compat --base <基準提交>`。路徑、operationId、型別名稱、授權、錯誤、nullable 與 deprecated 語義均須保持；生成碼變更需檢視，不能只因生成成功就認定相容。

OpenAPI 拆分不等同 runtime 完整驗證，也不代表資料庫交易的原子性。HTTP handler、業務規則與交易邊界另依實作核對。
