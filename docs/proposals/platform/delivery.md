# 實作契約與待決策清單

<a id="實作契約與待決策清單"></a>

- 狀態：本次基礎雛形採群組定位打卡、1:1 分群查本人紀錄；規則以功能契約為準，不把後續平台規劃當成首版規格。

- 順序依 ADR 0002、基礎設施 與 基礎出勤。群組是打卡範圍，MINI App 單次定位為打卡必經步驟；後台、通知、複雜權限及半徑判斷仍屬後續範圍。

  - [ADR 0002](../../decisions/0002-baseline.md)

  - [基礎設施](../../development/environment/verification.md)

  - [基礎出勤](../../domains/attendance/README.md)

<a id="第一個可實作切片"></a>

## 第一個可實作切片

[目前先完成 HTTP／容器／Cloud Run 骨架，再接入資料與首期 LINE 入口。 後續平台切片參考：接收 user/group Webhook、驗簽、持久接收、去重與 source 分流。 目前只有 最小接入試作，不含持久接收；不新增健康檢查路由。 本切片不執行任意 Postback、不授予群組管理權、不自動 Push；未知事件安全記錄中繼資料並忽略。 這個切片能先開始，不等於完整業務產品。](../../integrations/line/setup.md)

<a id="產品決策閘門"></a>

## 產品決策閘門

本次基礎雛形只依賴 P01、P04、P05、P06 及定位閘門 P08 中與群組定位打卡直接相關的決策。P02 群組 owner、P03 管理員與 P07 圖片屬後續平台能力，不是目前群組打卡、MINI App 定位或 1:1 查詢的前置條件；未來要加入管理後台或群組治理時才重新啟用。

| ID | 未決策問題 | 影響／禁止猜測 | 可繼續的獨立工作 |
| --- | --- | --- | --- |
| P01 | 已定為基礎出勤；入口、人員資格與時間歸屬待確認 | 不擴充排班、薪資或補登 | 執行與部署基礎 |
| P02 | 後續：群組 owner 如何驗證、綁定、移交與撤銷 | 不把首次點擊者升權；不影響目前由後端短效請求綁定群組的打卡流程 | 未來群組治理與拒絕式授權介面 |
| P03 | 後續：管理員登入來源與身分 provisioning | 不自動給 LINE 使用者管理權；目前雛形沒有管理員角色 | 未來 RBAC 契約與管理後台 |
| P04 | LINE Provider／channels、地區與 MINI App 資格 | 不假設跨 Provider user ID 相同 | token 驗證 adapter 的受控測試 |
| P05 | 訊息額度、保存期、通知與超時補發政策 | 不自行決定行銷、永久保存或結果不明後補發 | 狀態機、可配置退避 |
| P06 | Cloud Run 已定；區域、服務帳戶與 IAM 待配置，Worker 延後 | 不把擴充規劃當首版前置條件 | 本機啟動與容器設計 |
| P07 | 後續：圖片儲存、公開 URL 與刪除政策 | 與目前出勤雛形無關；未來仍不得把圖片 bytes 存進 Firestore | 模板純資料驗證 |

每筆決策需：owner、答案、狀態、日期、受影響 API/schema/tests；確認後以 ADR／功能契約落地。 不需要為每個一般技術選擇詢問；只有依賴缺失產品答案的實作保持 blocked。

<a id="平台交付階段參考"></a>

## 平台交付階段參考

狀態：規劃；不預建全部套件、不宣稱現有 HTTP 服務已實作以下功能。

<a id="當前交付順序"></a>

## 當前交付順序

- 依 ADR 0002、分階段落地 與 基礎出勤：執行／Cloud Run 骨架 → 資料與首期 LINE 接入 → 上班、下班與本人紀錄。

  - [ADR 0002](../../decisions/0002-baseline.md)

  - [分階段落地](../../development/environment/verification.md)

  - [基礎出勤](../../domains/attendance/README.md)

<a id="後續平台擴充參考"></a>

## 後續平台擴充參考

以下舊平台階段僅供未來需求確認後參考，不是目前的實作順序或必做清單。

| 階段 | 範圍 | 驗收 |
| --- | --- | --- |
| 1 | 一對一、Webhook、Inbox、Flex Reply | 原始 body 驗簽、多事件、空事件、重送、Unsend、無效輸入 |
| 2 | 群組事件、群組綁定、角色授權 | 缺少 actor、越權、退出重入、私人內容不回群 |
| 3 | Go MINI App 頁面、LIFF、session、API | 錯誤 channel、過期 token、跨使用者／群組、CSRF |
| 4 | Outbox、Worker、Garr | 重啟、租約過期、多副本競爭、429、超時結果不明、上限與背壓 |
| 5 | Rich Menu、Flex 管理 | payload／altText、不可變版本、測試發送、發布回復、裝置顯示 |
| 6 | 投產 | 持續 Worker 資源、IAM、索引、額度、保存／刪除政策、告警與復原 |

Service Message 等正式資格功能需另外驗收後啟用。

<a id="codex-任務輸入與完成條件"></a>

## Codex 任務輸入與完成條件

- 每次只讀 overview 路由所需文件，再讀當前範圍的原始碼。 交接列出目標、擁有檔案、依賴與驗收；一檔一寫入者。小任務不要自動展開全部角色。 新增 package 前依 layout 選位置，依 dependencies 核對 import 方向。 完成時列出實際變更、執行過的檢查、未實作部分與外部阻礙；Graphify/Repomix 的產出不是功能測試。

  - [overview](../../architecture/runtime/overview.md)

  - [layout](../../architecture/runtime/code-layout.md)

  - [dependencies](../../architecture/runtime/dependencies.md)
