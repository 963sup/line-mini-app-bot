# 目前目錄與責任

本頁描述目前的主要目錄；邊界規則見 [後端功能邊界](backend-boundaries.md)。

## 從需求定位修改範圍

先讀對應規格，再以 Graphify 定位當前符號並核對原始碼；下表是導航，不代表所列能力已完成環境驗收。命名依[命名規範](../../development/naming.md)，檢查依[驗收矩陣](../../agents/acceptance.md#變更驗收矩陣)。

| 需求 | 規格與不可破壞的契約 | 主要修改入口 |
| --- | --- | --- |
| 公開 API 格式或操作 | [API 工作流程](../../../api/README.md)、[契約索引](../../../api/contract/README.md) | `api/contract/` → 生成 Go／TS → 對應 `business/*/http` 與前端呼叫端 |
| 出勤、工時、更正與請假 | [出勤索引](../../features/attendance/README.md) | `backend/internal/business/attendance/`、`frontend/miniapp/src/business/attendance/` |
| Project 所有權與跨來源存取 | [獨立 Project](../../features/project/projects.md) | `backend/internal/business/project/`、`group/` 與來源讀取 adapter；前端 `business/project/` |
| Repository、物件、Issue、Discussion | [功能與資料規則](../../features/repository/README.md) | `backend/internal/business/` 對應功能、各自 HTTP／Firestore adapter；前端對應 `business/` |
| Session、登入與請求認證 | [Session 契約](../identity/browser-session.md)、[HTTP 契約](../api/http-contracts.md) | `backend/internal/identity/`、`frontend/miniapp/src/identity/` |
| LINE Webhook、導覽或訊息 | [LINE 接入](../../development/line-prototype.md) | `backend/internal/navigation/`、功能 `line/` adapter、`internal/server/` 組裝 |
| 摘要／天氣 | [摘要入口](../../../api/contract/summary/README.md)、[天氣入口](../../../api/contract/weather/README.md) | `backend/internal/business/summary/`、`weather/forecast/`；不自動加入公開 HTTP 契約 |
| Firestore 欄位、交易或遷移 | 所屬功能規格與[資料契約適用範圍](dependencies.md#契約與不變條件的維護位置) | 功能 `firestore/` adapter；索引 `deploy/firestore/`；既有遷移 `backend/cmd/` |
| 部署、設定或前端資產供應 | [Cloud Run 操作](../operations/cloud-run.md)、[前端生命週期](../frontend/react-frontend.md) | `Dockerfile`、`deploy/`、`backend/internal/server/`、`browser/`、前端建置設定 |

## 主要目錄

```text
backend/
  .env.example                 # 後端執行期設定範本
  go.mod、go.sum
  cmd/server/、attendancemigrate/、firestorecheck/、projectmigrate/、richmenu/
  internal/
    business/
      attendance/
        punch/、records/、worktime/、corrections/、approvals/、leave/、board/
        http/、firestore/、line/、presentation/  # 領域整合入口與交易
      summary/
        messagecollection/、messagecollection/line/
        summarization/         # 摘要用例與工作契約
          http/、line/、cloudtasks/、summarizer/
        line/                  # 收集與摘要指令的組合入口
        firestore/             # 訊息與工作共用版本／交易
      group/、repository/、project/
      object/、subobject/、objectimport/、issue/、discussion/、journal/
      weather/forecast/
        cwa/、line/
    identity/                  # 身分與驗證契約
      http/、line/、google/
    httpapi/                   # OpenAPI 路由接線
      generated/               # 生成路由、DTO、embedded spec
    browser/                   # Next.js 靜態頁、CSP、資源供應
    diagnostics/geolocation/   # 獨立定位示範
    navigation/line/、lineui/   # 導覽與選單
    transport/http/、line/     # 共用傳輸與格式
    persistence/firestorekey/  # 共用且相容的文件 ID 編碼
    architecture/              # 依賴方向與核心純度檢查
    server/                    # 設定、client、組裝、啟停
frontend/miniapp/
  src/pages/                   # Next.js Pages Router；靜態輸出
  src/navigation/home/
  src/business/attendance/     # 打卡、紀錄、更正、審核、請假
  src/business/repository/、project/、object/、subobject/、issue/、discussion/、journal/
  src/business/form/journal/   # Google 日誌表單入口
  src/identity/、platform/line/、shared/、shell/
  src/diagnostics/geolocation/
  tests/                       # React Strict Mode 與非同步行為檢查
  public/assets/、out/          # 產物，不提交
deploy/
  cloudbuild.yaml              # 手動／自訂 Trigger 建置與部署
  firestore/indexes.json       # Firestore 索引宣告
api/
  redocly.yaml                 # 契約 lint／bundle 規則
  contract/                    # 唯一入口與依功能拆分的契約
  client/                     # TS 生成型別與 openapi-fetch factory
  codegen/                    # 程式碼產生設定
  capabilities.json           # 功能清單
assets/line/richmenu/           # LINE 選單素材
scripts/
  tooling/                     # 環境、檢查、編輯器與前端資產工具
  analysis/                    # 圖譜與來源範圍稽核
  api/                         # 契約生成、功能清單與路由分派檢查
  release/                     # 發布驗證與狀態檢視
docs/、.github/、.vscode/、.agents/、.codex/
```

pnpm 管理 JS 工作區，Go module 位於 backend；Turborepo 協調產物與檢查。公開 URL 與既有資源 aliases 保留。Go binary 透過 FRONTEND_DIR 讀取獨立的 Next.js 靜態輸出；正式容器不需要 Next SSR 服務。

前端身分、請求取消、靜態資產與頁面生命週期見 [React 前端](../frontend/react-frontend.md)。

業務拆分與相容性驗證見 [目前模組與責任](layout.md)。
