# 目前目錄與責任

本頁描述目前原始碼配置；產品能力仍須區分本機實作與正式環境驗收。
未來的 direct/group/jobs/Garr 等完整配置移至 [未來平台規劃](future-platform.md)，有實際需求才建立。

```text
cmd/
  server/main.go          # 執行 HTTP 服務，回報致命錯誤
  attendancemigrate/      # 出勤查詢索引回填工具
  firestorecheck/         # 唯讀連線檢查
  richmenu/               # 選單管理 CLI
internal/
  server/                 # 服務生命週期、設定、資料庫與 adapter 組裝
  attendance/             # 出勤資料型別、儲存能力契約、更正規則與月報
  worktime/               # 純計算、事件套用與事件序列驗證
  summary/                # 摘要資料、介面與工作流程
  weather/                # 預報契約與目前唯一的 CWA 查詢實作
  lineui/                 # Flex、選單、群組名稱等 LINE／Web 共用呈現
  adapters/
    http/                 # MINI App、出勤 API、定位診斷與內部摘要 worker
    line/                 # LINE webhook、事件轉換、出勤／摘要／天氣回應與發送
    firestore/            # 出勤、啟動憑證、摘要工作持久化與查詢
    cloudtasks/           # 摘要工作入列
    summarizer/           # Gemini／OpenAI 模型與回應檢核
  web/
    embed.go              # 僅嵌入 runtime 頁面與資源
    pages/                # layout、home、history、correction、review、punch、geolocation
    static/js/            # session、api、time 與頁面控制
    static/css/           # app、geolocation
    tests/                # 既有瀏覽器腳本檢查；不嵌入執行檔
assets/richmenu/           # CLI 管理的圖片
scripts/                  # mise 任務實作、設定與分析工具
docs/                     # 功能、架構、開發、整合、維運及工具索引
```

## 歸屬規則

- `cmd/server` 只啟動 `server.Run`；`internal/server` 組裝外部實作與生命週期。
- HTTP 與 LINE 事件分屬 adapter，核心契約不 import adapter 或外部 SDK。
- 出勤與工時計算保留不同套件：出勤處理範圍與更正，worktime 處理時間與事件規則。
- 外部模型與佇列依賴由組裝點注入 summary 的 Model／Queue／Sender 介面。
- 小型 weather 暫留 CWA 實作，不為單一供應商增加空層。
- Web 檔名表達責任；pages 與 static 由 web 套件管理，HTTP adapter 提供路由。
- 公開頁面與 API 路徑維持原契約。舊 miniapp.js、attendance.js、app-api.js、miniapp.css 資源網址保留 alias；新頁使用 session.js、punch.js、api.js、app.css。
- 上游技能保留原位置與來源；產品檔案樹不重複展開技能內容。

命名見 [命名規則](../development/naming.md)，依賴見 [依賴規則](dependencies.md)。
