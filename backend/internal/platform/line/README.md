# LINE 平台能力

提供多個領域共用的 LINE Group Member API 查詢機制與 [Rich Menu 同步](richmenu/README.md)。由 bootstrap 注入既有 Membership port，不依賴 transport、identity 或 business。

## 入口與語意

[membership.go](membership.go) 每次查詢目前成員資料，不快取歷史互動作為授權證據。只有 HTTP 200、有效 JSON 且 userId 與查詢者一致才回傳 true。404 回傳 false；400、認證失敗、限流、網路錯誤與無效資料回傳 ErrMembershipUnavailable，均不得授權。404 也可能表示 Bot 不在群組，不能據此宣稱使用者已離群。

預設查詢上限一秒，Attendance 由 bootstrap 指定五秒；上游較短 deadline 優先。bootstrap 的 HTTP Client 拒絕 redirect。ValidGroupID 僅檢查 LINE 群組 ID 格式，不代表成員身分或業務權限。

## 責任邊界

Webhook 驗簽、事件派送與 Reply Token 語意保留在 transport/line；共用導航與 Rich Menu 定義保留在 transport/line/navigation；領域 Flex、Quick Reply 與政策保留在各領域 adapter。MINI App token 驗證與 Session 屬於 identity。直接使用官方 SDK 已提供的訊息型別，沒有新增通用訊息框架。

## 驗證與來源

在儲存庫根目錄執行 `bash scripts/tooling/tool-exec.sh go -C backend test ./internal/platform/line ./internal/architecture ./internal/bootstrap ./internal/transport/line/...`。測試不呼叫真實 LINE；正式渠道整合另行驗收。

LINE 官方 [Group Member Profile](https://developers.line.biz/en/reference/messaging-api/#get-group-member-profile) 的核對來源為 line-developers-docs-source main revision c15d6d3fa2eff7fcae65554c20be5b1e60f9a10a；SDK 保持 v8.23.0。

[上層導覽](../README.md) · [上層工作指引](../AGENTS.md)
