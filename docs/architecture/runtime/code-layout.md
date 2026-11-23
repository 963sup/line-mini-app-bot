# 目前目錄與責任

<a id="目前目錄與責任"></a>

[本頁描述目前的主要目錄；邊界規則見 後端功能邊界。](backend-boundaries.md)

<a id="從需求定位修改範圍"></a>

## 從需求定位修改範圍

- 先讀對應規格，再以 Graphify 定位當前符號並核對原始碼；下表是導航，不代表所列能力已完成環境驗收。命名依命名規範，檢查依驗收矩陣。

  - [命名規範](../../standards/naming.md)

  - [驗收矩陣](../../development/tooling/testing/commands.md#變更驗收矩陣)

- 需求：公開 API 格式或操作<br>規格與不可破壞的契約：API 工作流程、契約索引<br>主要修改入口：api/contract/ → 生成 Go／TS → 對應 business/*/http 與前端呼叫端

  - [API 工作流程](../../../api/)

  - [契約索引](../../../api/contract/)

- [需求：出勤、工時、更正與請假<br>規格與不可破壞的契約：出勤索引<br>主要修改入口：backend/internal/business/attendance/、frontend/miniapp/src/business/attendance/](../../domains/)

- [需求：Project 所有權與跨來源存取<br>規格與不可破壞的契約：獨立 Project<br>主要修改入口：backend/internal/business/project/、group/ 與來源讀取 adapter；前端 business/project/](../../domains/project/project.md)

- [需求：Repository、物件、Issue、Discussion<br>規格與不可破壞的契約：功能與資料規則<br>主要修改入口：backend/internal/business/ 對應功能、各自 HTTP／Firestore adapter；前端對應 business/](../../domains/)

- 需求：Session、登入與請求認證<br>規格與不可破壞的契約：Session 契約、HTTP 契約<br>主要修改入口：backend/internal/identity/、frontend/miniapp/src/identity/

  - [Session 契約](../../contracts/identity.md)

  - [HTTP 契約](../../contracts/http.md)

- [需求：LINE Webhook、導覽或訊息<br>規格與不可破壞的契約：LINE 接入<br>主要修改入口：backend/internal/transport/line/navigation/、功能 line/ adapter、internal/bootstrap/ 組裝](../../integrations/line/setup.md)

- 需求：摘要／天氣<br>規格與不可破壞的契約：摘要入口、天氣入口<br>主要修改入口：backend/internal/business/summary/、weather/forecast/；不自動加入公開 HTTP 契約

  - [摘要入口](../../../api/contract/summary/)

  - [天氣入口](../../../api/contract/weather/)

- [需求：Firestore 欄位、交易或遷移<br>規格與不可破壞的契約：所屬功能規格與資料契約適用範圍<br>主要修改入口：功能 firestore/ adapter；索引 firestore.indexes.json；已退役的舊 Project 遷移不再提供](dependencies.md#契約與不變條件的維護位置)

- 需求：部署、設定或前端資產供應<br>規格與不可破壞的契約：Cloud Run 操作、前端生命週期<br>主要修改入口：Dockerfile、cloudbuild.yaml、backend/internal/bootstrap/、browser/、前端建置設定

  - [Cloud Run 操作](../../operations/cloud-run.md)

  - [前端生命週期](react-lifecycle.md)

<a id="主要目錄"></a>

## 主要目錄

```text
.env.example                   # 根目錄執行期設定範本
backend/
  go.mod、go.sum
  main.go                    # 唯一執行入口
  internal/
    business/
      attendance/
        domain/、application/、adapter/  # 領域整合入口與交易
      summary/
        domain/、application/
        adapter/firestore/、http/、line/、messaging/、summarizer/
      group/、repository/、project/
      object/、subobject/、issue/、discussion/、journal/
      weather/domain/entity/forecast/
      weather/adapter/cwa/forecast/、weather/adapter/line/forecast/
    identity/                  # 身分與驗證契約
      http/、line/、google/
    transport/http/api/                   # OpenAPI 路由接線
      generated/               # 生成路由、DTO、embedded spec
    transport/http/browser/                   # Next.js 靜態頁、CSP、資源供應
    transport/http/geolocation/   # 獨立定位示範
    transport/line/navigation/   # 導覽與選單
    transport/http/、line/     # 共用傳輸與格式
    platform/firestore/key/  # 共用且相容的文件 ID 編碼
    platform/line/richmenu/ # LINE 自動選單同步
    architecture/              # 依賴方向與核心純度檢查
    bootstrap/                    # 設定、client、組裝、啟停
frontend/miniapp/
  src/pages/                   # Next.js Pages Router；靜態輸出
  src/navigation/home/
  src/business/attendance/     # 打卡、紀錄、更正、審核、請假
  src/business/repository/、project/、object/、subobject/、issue/、discussion/
  src/business/journal/        # 工程日誌填寫與查閱
  src/identity/、platform/line/、shared/、shell/
  src/diagnostics/geolocation/
  tests/                       # React Strict Mode 與非同步行為檢查
  public/assets/、out/          # 產物，不提交
cloudbuild.yaml              # 手動／自訂 Trigger 建置與部署
firestore.indexes.json       # Firestore 索引宣告
api/
  codegen/redocly.yaml                 # 契約 lint／bundle 規則
  contract/                    # 唯一入口與依功能拆分的契約
  client/                     # TS 生成型別與 openapi-fetch factory
  codegen/                    # 程式碼產生設定
  contract/capabilities.json           # 功能清單
assets/line/richmenu/           # LINE 選單素材
scripts/
  tooling/                     # 環境、檢查、編輯器與前端資產工具
  analysis/                    # 圖譜與來源範圍稽核
  api/                         # 契約生成、功能清單與路由分派檢查
  release/                     # 發布驗證與狀態檢視
docs/、.github/、.vscode/、.agents/、.codex/
```

- pnpm 管理 JS 工作區，Go module 位於 backend；Turborepo 協調產物與檢查。公開 URL 與既有資源 aliases 保留。Go binary 透過 FRONTEND_DIR 讀取獨立的 Next.js 靜態輸出；正式容器不需要 Next SSR 服務。

- [前端身分、請求取消、靜態資產與頁面生命週期見 React 前端。](react-lifecycle.md)

- [業務拆分與相容性驗證見 目前模組與責任。](code-layout.md)
