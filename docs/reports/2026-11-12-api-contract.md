# API 契約能力歸屬（2026-11-12）

原結構只拆分 Attendance，Identity／Session 與 Repository 仍內嵌根契約；出勤來源反向引用根入口的 APIError，Corrections 與 Approvals 互相引用。這次把來源歸屬與依賴方向整理一致，完整索引及維護規則見 [契約能力索引](../../api/contract/README.md)。

- Identity 擁有設定、本人與 Session 的操作及 DTO；Repository 擁有容器列表、詳情與 DTO。
- common 僅保存共用錯誤與認證方式；openapi.yaml 保留 metadata、路徑與 components 引用索引。能力檔不再反向引用根入口。
- 補卡列表與審核共用的 CorrectionReview、撤回與審核共用的 CorrectionDecisionResult 移到 Corrections，保留型別名稱與內容；Approvals 單向引用 Corrections。
- Summary／Weather 維持既有事件／內部工作契約；共用 client、codegen、compatibility 不按能力複製。

本次僅整理契約來源及文件，不改 HTTP handler、權限、產品規則或資料交易。handler 參數傳遞與 runtime 驗證的統一是另一項行為調整，不能由契約搬移推定已完成。

查證沿用本次 Context7 查得的 Redocly v2 多檔 `$ref`／bundle 文件，本機固定版本以 package.json 為準。Graphify 用於程式關係定位與更新，不能證明所有契約引用；Repomix 限定打包契約、生成腳本、HTTP 接線與相關文件，另以完整解析結構比較確認公開契約語義。

實際驗證：拆分前後完整 dereferenced JSON 相等；Go／TypeScript 重新生成後與原檔完全相同；`pnpm api:check`、`pnpm api:compat --base d395e0023cb719ee5622ecf1b8ed54c9d696ec0f`、設定檢查、變更 Markdown 的本機檔案連結與 `git diff --check` 均通過。程式與文件圖譜均已重建；文件圖譜未解析本機連結為 0。未修改應用行為，因此未重跑完整 Go／前端測試或外部整合驗收。
