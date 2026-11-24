# 0004：前後端與 API 契約分離的 Monorepo

- 狀態：Accepted
- 日期：2026-11-10

## 背景

專案需要共享 HTTP 契約，同時讓 Go 後端與 Next.js MINI App 保持各自清楚的建置與責任。

## 決策

- 根目錄分為 `backend/`、`frontend/` 與 `api/`。
- 前端使用 Next.js Pages Router、React 與 TypeScript 靜態輸出，不新增 SSR Node runtime。
- Go service 同來源提供靜態頁面與 API。
- OpenAPI 產生 Go 與 TypeScript 契約型別；產生檔不得手改。
- pnpm workspace 管理 JavaScript 相依，鎖檔提交版本控制。

## 後果

API 變更必須先改契約並同步兩端。正式映像只有 Go runtime，但建置階段需要 Node 工具鏈。前端不得使用公開環境變數承載秘密；真實 LINE 裝置與正式設定需獨立驗收。
