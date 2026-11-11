# 日誌表單

日誌分類為 `business/form/journal`。群組快速選單在「打卡」旁提供「日誌」，透過既有 MINI App URL 的 `journal=1` 開啟日誌頁；Help 選單及 MINI App 首頁也提供入口。

## 操作契約

| 欄位 | 內容 |
| --- | --- |
| Actor / Resource / Action | 使用者開啟指定 Google 日誌表單；此入口不驗證或宣稱填寫人的 LINE 身分 |
| Context / Authority | 公開頁面入口；Google 表單擁有者設定填寫、Google 登入及上傳權限 |
| Policy / Decision | 僅嵌入固定表單；不由 URL 接受任意表單網址；Google 決定是否接受提交 |
| Result | 顯示表單；提交結果以 Google 表單為準，資料由 Google 保存；不更新出勤或本系統資料 |
| 重送／併發 | 重開頁面不送出表單；重複提交規則由 Google 表單設定負責 |
| Event / Audit | 不傳送 LINE token、群組或使用者資料給表單；本系統不記錄提交事件，也不發送完成通知 |

## 實作位置

- [日誌畫面](../../frontend/miniapp/src/business/form/journal/View.tsx)、[頁面入口](../../frontend/miniapp/src/pages/miniapp/journal.tsx)。
- [群組選單](../../backend/internal/navigation/lineui/menu.go)、[頁面路由及 CSP](../../backend/internal/browser/routes.go)。
- [共用版型](../../frontend/miniapp/src/shell/Shell.tsx)：識別日誌 LIFF 參數；日誌不顯示出勤身分操作，登入狀態變更不重建表單。

沒有新增公開業務 API、後端業務套件或 Firestore 儲存層。
Google Forms 嵌入只在日誌頁允許 `https://docs.google.com`，保留其他頁面的既有 CSP。
若登入、上傳或嵌入受限，使用畫面的「開啟表單」連結；不把 iframe 載入當成送出成功。

## 驗收限制

需以 LINE iOS、Android 與外部瀏覽器確認表單顯示、Google 登入、上傳（若表單有要求）及提交。
目前環境無法讀取 Google 表單或直接開啟 LINE 官方網站；表單權限與實機相容性待驗收。
LINE URL 參數流程依官方來源 main 修訂 `dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54` 的 LIFF opening 文件核對。
Next.js 靜態頁依 Context7 `/vercel/next.js/v16.4.5` 與本機相同版本指引核對。
