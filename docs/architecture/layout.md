# 目前目錄與責任

本頁描述第四階段完成後的配置；邊界規則見 [後端功能邊界](backend-boundaries.md)。

```text
backend/
  go.mod、go.sum
  cmd/server/、attendancemigrate/、firestorecheck/、richmenu/
  internal/
    attendance/                # 純領域模型、規則與能力契約
      worktime/                # 時間與事件計算
      http/                    # 出勤 API、DTO、契約驗證
      firestore/               # 文件格式、交易、查詢、回填與錯誤轉譯
      line/                    # 聊天指令與出勤卡片
      presentation/            # 不依賴 LINE SDK 的共用呈現
    summary/                   # 摘要契約與用例
      http/、firestore/、line/、cloudtasks/、summarizer/
    weather/                   # 預報契約
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
  src/features/attendance/     # 打卡、紀錄、更正、審核
  src/identity/、platform/line/、shared/、shell/
  src/diagnostics/geolocation/
  tests/                       # React Strict Mode 與非同步行為檢查
  public/assets/、out/          # 產物，不提交
api/
  contract/                    # 唯一入口與出勤契約，共 14 個 operations
  client/                     # TS 生成型別與 openapi-fetch factory
  codegen/、compatibility/
assets/richmenu/
docs/、scripts/、.github/、.agents/、.codex/
```

pnpm 管理 JS 工作區，Go module 位於 backend；Turborepo 協調產物與檢查。公開 URL 與既有資源 aliases 保留。Go binary 透過 FRONTEND_DIR 讀取獨立的 Next.js 靜態輸出；正式容器不需要 Next SSR 服務。

前端已使用 React hooks/state、身分 Context 與具取消／世代檢查的請求；Strict Mode 已恢復，舊控制器及事件橋接已移除。舊 JS 資源 URL 僅回傳無操作的相容內容，不被新頁面載入。功能生命週期見 [React 前端](react-frontend.md)。客戶端導覽尚未導入。上游技能保留原位置，不重複加入產品功能樹。
