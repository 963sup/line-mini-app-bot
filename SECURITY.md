# Security policy

## Scope and supported versions

專案目前尚未宣告穩定發布或 LTS 支援範圍。安全修正以維護者目前開發版本為目標；不要將 work/main 分支存在視為已完成正式安全驗收。
本政策涵蓋本儲存庫程式、設定與憑證處理；第三方服務漏洞請依其官方通報流程處理。

## Report a vulnerability

請勿在公開 Issue、PR、日誌或聊天貼出 secret、token、私人位置、LINE 個資或可利用的完整細節。
若儲存庫已啟用 GitHub private vulnerability reporting，請使用：

https://github.com/963sup/line-mini-app-bot/security/advisories/new

本文件不代表該功能已啟用。若連結不可用，請先透過維護者 GitHub 公開個人頁面提供的聯絡方式要求私密通報管道；未取得管道前不要公開漏洞細節。
不宣告不存在的安全信箱、回覆期限、獎金或修復 SLA。

## Useful report contents

- 受影響 commit／版本、元件與必要環境。
- 最小重現步驟、預期／實際結果與影響範圍。
- 已遮蔽敏感內容的證據；請使用測試帳號與測試資料。
- 建議修復或緩解方式（若有）。不要測試未獲授權的外部服務或他人資料。

## Credential exposure

發現憑證外洩時先撤銷／輪替，檢查權限與使用紀錄，再處理歷史資料；刪除 Git 檔案不足以撤銷 token。
LINE secrets 與 Google 認證限安全注入，不提交版本庫或分析產出。

## Review boundaries

Firestore server SDK 使用 IAM 並繞過 Security Rules；後端仍需個人／群組資源授權。
瀏覽器座標、postback data 與前端 profile 都是不可信輸入。
第三方 skills/plugins 與更新需審查來源及權限；安裝不授權部署或外部操作。
安全掃描與測試不是無漏洞保證；未執行的真實渠道／裝置驗收必須明確列出。

## Analysis artifacts and Cloud development

Graphify 圖譜、Repomix 原文包、coverage／profile 與工具日誌可能包含程式、設定或個人資料；
僅在受控工作區處理，分享前檢查實際內容，不將產物提交至 Git 或公開通報。
`.env.example` 只能放非機密範例；真實憑證透過環境安全注入。
ignore 只控制各工具的收錄，不撤銷已提交的秘密；已追蹤檔案也不受 `.gitignore` 保護。
分析入口排除 `.agents`、`.claude`、`agent` 與 `skills-lock.json`，但排除規則及安全掃描
都不能證明剩餘內容不含敏感資訊。保留既有安全掃描，不以停用排除規則解決漏檔。
Codex Cloud 安裝、MCP 握手、應用測試與真實服務權限需分開驗證。
