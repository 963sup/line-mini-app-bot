# Membership 成員資格

Membership 管理成員關係與資格判定。Organization、Enterprise 和 Team 各自擁有資源狀態與管理歸屬。[組裝入口](../../../backend/internal/bootstrap/server.go)將成員資格查詢接到 Team 與 Enterprise，並註冊 Membership 的 HTTP handler。

## 現行讀取與授權

- Team 的空值或 `legacy` 模式沿用既有 LINE 群組資格。`membership` 模式只採獨立關係。未知模式、停用的 Team 或缺少有效資格時拒絕存取。團隊管理歸屬和候選來源見 [Team 規格](../team/design.md#獨立-membership-的讀取邊界)。
- 獨立模式須有有效的 Team `member` 或 `maintainer` 關係。成員關係記錄的上層類型、資源 ID 與世代須符合管理歸屬指定的有效 Organization 或 Enterprise 資格。查詢沿用呼叫端的 Firestore 交易。上層撤權使衍生 Team 資格失效，重新加入後也須重新授予。[資格讀取實作](../../../backend/internal/business/membership/adapter/firestore/qualification.go) · [世代判定](../../../backend/internal/business/membership/domain/qualification.go)
- Organization 的 Bot 私聊邀請及現行成員操作沿用既有 Organization 授權模式。[LINE 邀請入口](../../../backend/internal/business/membership/adapter/line/invitations.go) · [HTTP 成員入口](../../../backend/internal/business/membership/adapter/http/members.go) · [組織邀請操作](../../proposals/membership.md#首版-bot-私聊入口)

## 切換與驗收

獨立資格的全面切換仍須完成各資源的正向讀寫、入口與環境驗收。切換順序、儲存遷移與驗收條件由 [Membership 提案](../../proposals/membership.md#儲存與遷移方案)維護。
