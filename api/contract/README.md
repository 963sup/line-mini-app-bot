# 契約能力索引

採用 **按領域拆分 OpenAPI 契約（modular OpenAPI）**。各領域維護自己負責的操作與 schema，公開入口 `openapi.yaml` 與內部入口 `internal-openapi.yaml` 透過 `$ref` 組合。依操作責任決定契約歸屬，調整來源檔後透過既有流程產生 Go 與 TypeScript 程式碼。 **契約按需分檔。** 小型領域可用一份規格承載相關操作與 schema，有明確的能力邊界或獨立維護需求時再拆分。

維護本能力的操作與 schema；根規格負責組合，公開格式由契約產生。

## 閱讀與修改入口

先查閱[能力與實作清單](capabilities.json)、[internal-openapi.yaml](internal-openapi.yaml)、[openapi.yaml](openapi.yaml)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[Project規格](../../docs/domains/project/project.md) · [標籤](../../docs/domains/repository/labels.md)

本目錄以能力歸屬組織契約。公開 HTTP、事件與內部工作入口有不同形式，不能因歸類而新增公開 API。功能授權與業務規則仍由對應功能文件定義。

| 能力／責任 | 來源 | 範圍 |
| --- | --- | --- |
| 公開 HTTP 組合入口 | [openapi.yaml](openapi.yaml) | API metadata、既有路徑與 components 的 `$ref` 索引 |
| 共用傳輸契約 | [common/transport.yaml](common/transport.yaml) | APIError、Bearer 與 Cookie security schemes；不承載業務模型 |
| PDF 工具 | [operations/poppler.yaml](operations/poppler.yaml) | 已登入使用者的 PDF 文字擷取測試，不儲存資料 |
| Identity | [identity/identity.yaml](identity/identity.yaml) | MINI App 公開設定、本人、Session 建立／續期／撤銷及其 DTO |
| Repository | [repository/repositories.yaml](repository/repositories.yaml) | 容器列表、詳情、分頁及容器 DTO；不提供手動建立或管理 API |
| Enterprise | [enterprises.yaml](enterprise/enterprises.yaml)、[types.yaml](enterprise/types.yaml) | 企業開通、詳情、Owner 與組織連結 |
| Organization | [context.yaml](organization/context.yaml) | 組織選取、建立及 Owner 移轉 |
| Team | [teams.yaml](team/teams.yaml) | 團隊管理、組織團隊查詢與企業團隊指派 |
| Group | [group/groups.yaml](group/groups.yaml) | 工程群組列表、詳情與當下成員資格 |
| Repository / Space | [repository/spaces.yaml](repository/spaces.yaml) | 庫內工區，沿用所屬 Repository 的授權 |
| Project | [project/owner-projects.yaml](project/owner-projects.yaml) | 個人或工程群組擁有的專案與items，來源唯讀；[Project規格](../../docs/domains/project/project.md) |
| Project legacy | [project/projects.yaml](project/projects.yaml)、[items.yaml](project/items.yaml) | 原Repository-scoped路由與DTO相容轉接；同一用例，不構成新擁有權 |
| Repository / Labels | [repository/labels.yaml](repository/labels.yaml) | 標籤目錄列出與建立，規格見 [標籤](../../docs/domains/repository/labels.md) |
| Issue | [issue/issues.yaml](issue/issues.yaml) | 空間事項列表、詳情、建立及建立者修改；操作規格見 [事項](../../docs/domains/issue/design.md) |
| Discussion | [discussion/discussions.yaml](discussion/discussions.yaml)、[comments.yaml](discussion/comments.yaml) | 主題讀寫及單層留言；作者編輯，規格見 [討論](../../docs/domains/discussion/design.md) |
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
| PDF 工具 | [operations/poppler.yaml](operations/poppler.yaml) | 已登入使用者的 PDF 文字擷取測試，不儲存資料 |
| Identity | [AGENTS.md](identity/AGENTS.md) | 既有公開契約 |
| Repository | [AGENTS.md](repository/AGENTS.md) | Bot 管理的群組空間 |
| Issue | [AGENTS.md](issue/AGENTS.md) | 基礎事項 API；types／comments 仍為雛形 |
| Discussion | [AGENTS.md](discussion/AGENTS.md) | 主題／單層留言已接入；categories 仍為雛形 |
| Project | [AGENTS.md](project/AGENTS.md) | Project／items 已接入；fields／views 為雛形 |
| Attendance | [AGENTS.md](attendance/AGENTS.md) | 既有出勤契約 |
| Summary | [AGENTS.md](summary/AGENTS.md) | 事件與內部工作契約 |
| Weather | [AGENTS.md](weather/AGENTS.md) | 既有事件入口契約 |

## 歸屬與依賴

命名遵循 [業務實體、集合與 API 命名](../../docs/standards/naming.md#業務實體集合與-api-命名)：單一實體型別用單數，資源集合及 REST 集合路徑用複數。既有單例及公開契約相容例外按該規範處理；能力名稱不以英文單複數機械判定。

- 根入口只組合來源，不內嵌能力的 path、schema 或 security scheme 定義。路徑與原有 component 名稱仍保留在入口索引，維持公開介面。
- 能力來源不得引用 `openapi.yaml`；共用錯誤與認證定義由 common 擁有。security requirement 名稱由根入口的 securitySchemes 組合解析，能力檔不是獨立完整 OpenAPI 文件。
- 能力擁有自己的操作、輸入及結果；真正跨能力共用的業務資料引用其擁有者，不能全部搬到 common。
- 出勤依賴方向：Approvals → Corrections → Records → Worktime；Approvals 亦直接引用 Records，Punch／Leave 引用 common。Corrections 不依賴 Approvals。
- `CorrectionReview` 名稱保持相容，但內容是列表與審核共用的補卡申請快照，因此由 Corrections 擁有；`CorrectionDecisionResult` 供撤回與審核使用，同樣由申請生命週期擁有者保存。
- 單一能力檔可包含相關 paths 與 schemas，不要求每個 endpoint 一個檔案，也不為沒有公開入口的能力建立空 HTTP 契約。

## 修改與驗證

### 最小契約雛形

Repository milestones／statistics 與 Project／items 已依 [里程碑](../../docs/domains/repository/milestones.md)、[Project](../../docs/domains/project/project.md)與[洞察（Insights）](../../docs/domains/repository/insights.md) 接入。以下檔案目前以 `x-status: scaffold` 標記，只有空的 `paths` 與 `components.schemas`，尚未定義端點或資料欄位，也不由公開／內部組合入口引用。能力清單登記其擁有者，不表示功能已實作。

- Issue：[types.yaml](issue/types.yaml)（IssueType）、[comments.yaml](issue/comments.yaml)。
- Discussion：[categories.yaml](discussion/categories.yaml)。
- Project：[fields.yaml](project/fields.yaml)、[views.yaml](project/views.yaml)；projects／items 已接入。

以上是能力契約片段，不是可獨立啟動或呼叫的完整 OpenAPI 文件。實作時先補功能操作規格，再填入對應 paths／schemas、接入組合入口、更新能力清單與生成碼，並驗證相容性；既有 Repository 仍由 Bot 建立。

先改來源契約，再依 [API 工作流程](../README.md) 生成共用 Go／TypeScript 產物。Redocly 從根入口 bundle，不需每個能力各建 client 或生成器；相容性 baseline 不隨搬移刷新。

純歸屬重組先比較搬移前後 `redocly bundle --dereferenced` 的完整 JSON 結構，再執行 `pnpm api:check`。路徑、operationId、型別名稱、授權、錯誤、nullable 與 deprecated 語義均須保持；生成碼變更需檢視，不能只因生成成功就認定相容。

OpenAPI 拆分不等同 runtime 完整驗證，也不代表資料庫交易的原子性。HTTP handler、業務規則與交易邊界另依實作核對。

## 工作日誌

[Journal 契約](journal/journals.yaml) 定義發布、列表與單篇讀取。Feed 與 Dashboard 使用同一資源；兩者的查詢方式與正式啟用條件見 [日誌設計](../../docs/domains/journal/journal.md)。

[現行實作圖：基礎入口](../../docs/architecture/contracts/http.md#基礎入口) · [未來設計圖](../../docs/diagrams.md#未來設計)

## 下層文件入口

- [group](group/README.md)
- [journal](journal/README.md)
- [object](object/README.md)
