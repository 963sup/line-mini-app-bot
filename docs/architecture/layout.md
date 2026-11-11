# 目前目錄與責任

本頁描述第四階段完成後的配置；邊界規則見 [後端功能邊界](backend-boundaries.md)。

```text
backend/
  .env.example                 # 後端執行期設定範本
  go.mod、go.sum
  cmd/server/、attendancemigrate/、firestorecheck/、richmenu/
  internal/
    business/
      attendance/
        punch/、records/、worktime/、corrections/、approvals/、leave/、board/
        http/、firestore/、line/、presentation/  # 領域整合入口與交易
      summary/
        messagecollection/、messagecollection/line/
        summarization/         # 摘要用例與工作契約
          http/、line/、cloudtasks/、summarizer/
        firestore/             # 訊息與工作共用版本／交易
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
  src/business/attendance/     # 打卡、紀錄、更正、審核
  src/identity/、platform/line/、shared/、shell/
  src/diagnostics/geolocation/
  tests/                       # React Strict Mode 與非同步行為檢查
  public/assets/、out/          # 產物，不提交
deploy/
  cloudbuild.yaml              # 手動／自訂 Trigger 建置與部署
  firestore/indexes.json       # Firestore 索引宣告
api/
  redocly.yaml                 # 契約 lint／bundle 規則
  contract/                    # 唯一入口與依出勤子功能拆分的契約
  client/                     # TS 生成型別與 openapi-fetch factory
  codegen/、compatibility/
assets/line/richmenu/           # LINE 選單素材
scripts/
  tooling/                     # 環境、檢查、編輯器與前端資產工具
  analysis/                    # 圖譜與來源範圍稽核
  api/                         # 契約生成與相容性檢查
  release/                     # 發布驗證與狀態檢視
docs/、.github/、.vscode/、.agents/、.codex/
```

pnpm 管理 JS 工作區，Go module 位於 backend；Turborepo 協調產物與檢查。公開 URL 與既有資源 aliases 保留。Go binary 透過 FRONTEND_DIR 讀取獨立的 Next.js 靜態輸出；正式容器不需要 Next SSR 服務。

前端已使用 React hooks/state、身分 Context 與具取消／世代檢查的請求；Strict Mode 已恢復，舊控制器及事件橋接已移除。舊 JS 資源 URL 僅回傳無操作的相容內容，不被新頁面載入。功能生命週期見 [React 前端](react-frontend.md)。客戶端導覽尚未導入。上游技能保留原位置，不重複加入產品功能樹。

業務拆分與相容性驗證見 [Business 重構](../development/business-refactor.md)。
