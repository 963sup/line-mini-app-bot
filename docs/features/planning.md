# 里程碑、專案與總覽統計基礎雛形

狀態：本機基礎雛形已完成。依本次使用者授權接入里程碑、專案與真實統計；完成後停止擴充，正式整合驗收另計。

## 操作契約

| 欄位 | 里程碑 | 專案與收錄事項 | 統計 |
| --- | --- | --- | --- |
| Actor / Resource / Action | 可信 Session／Bearer 本人，同空間 Milestone；列出、建立、詳情、更新 | 同左；Project 與 ProjectItem；列出、建立、詳情、更新、加入／移除事項 | 同空間完整統計，只讀 |
| Context / Authority | 啟用空間與當下群組成員；建立者才能編輯、完成／重開 | 成員可讀／建立；專案建立者管理描述及收錄項目，不授予事項修改權 | 當下成員可讀 |
| Policy / Decision | 標題 1–120、描述最多 8000 字；可選曆日 YYYY-MM-DD；open／completed 手動切換 | 僅同空間既有事項；同事項可在不同專案，同專案不重複；ProjectItem ID 與 Issue ID 分開；沒有獨立專案狀態 | 事項三狀態、討論、里程碑、專案數；讀取完整集合，不以分頁筆數作總數 |
| Result | 內容、版本與更新時間；事項作者可選擇／清除一個同空間里程碑 | 專案內容、項目關聯、專案版本；來源事項完全不改動 | 真實數量及觀察時間；失敗顯示錯誤，不偽裝為 0 |
| 重送／併發 | 建立鍵重送同內容同 ID，異內容 409；更新 expectedVersion，衝突 409；無變化不加版本 | 建立同左；加入／移除以專案 expectedVersion 序列化；同事項衍生穩定獨立項目 ID；結果不明先重新讀取再決定，避免重複收錄 | 即時查詢無快取或背景工作；不提供跨請求的統計快照保證 |
| Event / Audit | 建立者／建立時間、更新操作者／版本／時間；無外部通知 | 同左，項目變動與專案版本 metadata 同交易 | 無寫入 |

里程碑完成是建立者手動確認；已完成事項比例只是進度，100% 不會自動完成里程碑。無事項時顯示「尚無事項」。專案顯示來源事項三狀態；首版不含跨空間、草稿、自訂欄位、拖曳、排程、刪除或圖表。里程碑及專案先不使用共用資源編號，不更動既有事項／討論收據。

## 分層及保存

里程碑核心位於 `business/repository/milestones`，Firestore 實作使用既有 Repository Store；專案核心與 adapter 位於 `business/project`，server 注入同一 Firestore client。統計是 `business/repository/statistics` 只讀用例；不建立通用引擎。

資料在既有 channel／repository 下的 `milestones`、`projects`；建立收據各自在對應功能的 receipt collection。專案項目位於 `projects/{id}/items/{itemId}`，只保存來源 issueId 與建立 metadata。更新交易檢查容器啟用／群組綁定、作者、版本及同空間引用，全部讀取先於寫入。

事項增加可選 milestoneId；舊資料缺少時視為未歸屬，更新省略保留、null 清除。建立事項收據雜湊不變。里程碑／專案進度由當前來源事項計算，沒有複製事項狀態；首版即時查詢成本隨資料量增長，後續依實際用量評估聚合，先不建立計數器或排程。

契約先修改原有 milestones／projects 骨架，再生成 Go／TS 型別；新增 statistics 契約並維護能力清單。正式 UI 組合既有元件，不修改 shadcn primitives。沿用原讀寫取消、身分隔離、防連點及重送模式。

## 驗證與後續

API 52 項操作、25 個契約來源、5 個事件的擁有者／命名／接線檢查、生成同步及向後相容通過，未刷新 baseline。Go 格式、lint、vet、完整既有 race suite、build 與前端型別、23 個既有測試、靜態輸出通過。設定及文件本機連結檢查通過。

隔離 Firestore emulator 暫存 overlay race smoke 通過：里程碑建立／重送／同鍵異內容拒絕、曆日驗證、作者編輯及手動完成／重開；事項里程碑引用、跨空間拒絕、省略保留／null 清除與舊標籤相容；專案作者權限、唯一獨立項目 ID、同事項跨專案、移出保留來源、來源事項即時狀態與進度；超過二十筆的完整統計、分頁游標綁定、建立及更新併發、停用容器交易拒絕。新增資源不消耗原共用編號。

暫存 React smoke 使用正式空間頁面確認統計失敗不顯示舊數字、里程碑建立連點鎖與原鍵重試、到期日、手動完成／重開、作者限制；事項歸屬保存／清除；專案收錄／移出、版本衝突後重新讀取與權限；離開總覽後丟棄遲到統計。驗證程式與 Java runtime 保留於 `/tmp`，不新增正式測試框架。

查證沿用 Context7 `/googleapis/google-cloud-go` 官方交易來源與本機 Firestore v1.29.2：回呼可重試、全部讀取先於寫入、read-only transaction 的 Documents／GetAll 共用快照。聚合 API 查詢未取得匹配文件，本次使用已查證的完整集合讀取，不宣稱已驗證聚合 API。Google skills／Developer Knowledge MCP 無可用入口；沒有操作正式資源。UI 依固定 shadcn 4.23.0 CLI 取得的官方 Radix Input／NativeSelect 範例組合，沒有改 primitives 或安裝新套件。govulncheck 未發現應用可達漏洞，另有三筆 required-module 未被應用呼叫的項目。

Graphify CLI 與 Repomix CLI 用於關係及原文交叉核對；圖譜未支援的檔案與動態呼叫不作正確性證據。最終原文快照在本機 `/tmp/planning-review.xml`，保留安全掃描並納入未追蹤新檔。修改保留本機，未提交、推送或部署。

完成後優先驗收真實 LINE 身分、群組、裝置及正式資料庫，再收集使用回饋。資料量擴大時評估查詢成本與有官方證據的聚合方案；不自動加入看板、自訂欄位、排程或趨勢圖。
