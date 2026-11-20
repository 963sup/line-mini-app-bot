# 獨立日誌feed-與-dashboard

<a id="獨立日誌feed-與-dashboard"></a>

```mermaid
flowchart LR
  h["獨立日誌、Feed 與 Dashboard"]
  h --- n0["日誌是獨立業務。每篇日誌指定一個發布群組。<br/>Feed 提供填寫與瀏覽入口，Dashboard 依日期與<br/>群組讀取同一批日誌，以每個有資料的群組一張圖<br/>表呈現。"]
```

<a id="日誌內容"></a>

```mermaid
flowchart LR
  h["日誌內容"]
  h --- n0["範圍：表頭<br/>欄位與語意：日誌識別碼、工作日期、發布群組、<br/>填寫人。允許同一天多篇紀錄。"]
  h --- n1["範圍：人員配置<br/>欄位與語意：管理職、管制人員、施作人員各有一<br/>份當日姓名清單，允許多人及兼任。職務只描述現<br/>場配置。"]
  h --- n2["範圍：今日工作<br/>欄位與語意：每列保存名稱與累計實際完成度（0–<br/>100%）。"]
  h --- n3["範圍：明日計畫<br/>欄位與語意：每列保存名稱與預計隔日累計完成度<br/>（0–100%）。"]
  h --- n4["範圍：工作關聯<br/>欄位與語意：每列可選一個儲存庫或專案，以類型<br/>及識別碼引用。關聯只屬於該列。"]
```

## 工作對象與實際作業空間

每筆工作可直接手寫。選擇儲存庫後，可搜尋該庫的物件或子物件，並選擇最多 20 個實際作業空間。子物件顯示「父物件 › 子物件」，選取時帶入父物件的空間；使用者可移除或加選。切換儲存庫會清除該筆引用，保留工作名稱與完成度。專案關聯維持原有行為，不提供庫內選取。

缺少空間時可輸入名稱並「建立並選用」。空間立即保存，即使取消日誌仍會保留；結果未知時鎖住該筆的關聯與移除，使用原鍵重試，完成前不能取消填寫或發布。

| 操作契約 | 規則 |
| --- | --- |
| Actor／Resource／Action | 可信 Session 使用者，在發布群組新增日誌；每列可引用一個工作對象及多個空間。 |
| Authority／Policy | 沿用群組成員與來源閱讀權；只有儲存庫關聯可帶入物件／空間，子物件須屬於指定父物件，所有項目須來自指定庫。 |
| Context／Decision | 發布交易讀取所選資料與儲存庫狀態；缺少來源或跨庫引用拒絕整篇發布，不新增日誌及收據。 |
| Result | 保存選取的 ID 與當時的工作對象名稱、空間名稱。日後改名或移動不改寫歷史；來源失權時連同這些欄位一併隱藏。日誌不更新來源進度。 |
| 重送／併發 | 原發布鍵與輸入內容決定收據；重試回原日誌，不重新計算名稱快照。空間建立沿用既有防重 API。 |
| Audit／實作 | 日誌保留作者及時間；[後端交易](../../../backend/internal/business/journal/adapter/firestore/store.go)、[前端選取](../../../frontend/miniapp/src/business/journal/WorkContext.tsx)及[API 契約](../../../api/contract/journal/journals.yaml)。 |

驗證沿用既有前端、Go 與 API 檢查。正式 Firestore 與 LINE 裝置驗收另行記錄。

<a id="畫面與雛型範圍"></a>

```mermaid
flowchart LR
  h["畫面與雛型範圍"]
  h --- n0["側邊欄的 Feed 依工作日期顯示日誌卡片，支援群<br/>組篩選。Dashboard 依工作日期與群組篩選，以<br/>shadcn 長條圖逐項呈現今日實際及明日預計累計<br/>完成度，不合計或平均。展開日誌明細可查閱人員<br/>配置並返回 Feed 原文。"]
  h --- n1["從 MINI App 側欄「快捷入口」點選「寫日誌」，<br/>登入後選擇發布群組，接著填寫及發布。Feed 的群組篩選僅供瀏覽。<br/>Feed 與 Dashboard 透過共用 API 讀取保存的日<br/>誌，逐列關聯從可讀儲存庫或群組專案選取。"]
```

<a id="正式群組發布的操作契約"></a>

```mermaid
flowchart LR
  h["正式群組發布的操作契約"]
  h --- n0["後端依可信 Session 設定填寫人，新增時核驗當<br/>下群組成員資格及每列來源讀取權。列表與詳情逐<br/>次依群組授權，來源失權時隱藏來源內容。寫入須<br/>驗證日期、非空工作名稱、百分比範圍與同次提交<br/>的防重識別。成功後只新增一篇日誌。關聯不授予<br/>閱讀權、不改來源進度。"]
```

<a id="資源名稱與-api-骨架"></a>

```mermaid
flowchart LR
  h["資源名稱與 API 骨架"]
  h --- n0["Journal 表示一篇工作日誌，集合路徑使用 <br/>/api/v1/journals。Feed 是按時間排列的瀏覽與<br/>發布入口，Dashboard 是依群組及工作日期查閱的<br/>畫面。兩者共用列表及詳情契約。"]
```

```mermaid
flowchart LR
  h["資源名稱與 API 骨架"]
  h --- n0["操作：POST /api/v1/journals<br/>用途：建立日誌，必填 Idempotency-Key。同一作<br/>者與同鍵同內容重送取得原結果，異內容回 409。"]
  h --- n1["操作：GET /api/v1/journals<br/>用途：Feed 可跨本人可讀群組，Dashboard 以 <br/>groupId 和 workDate 篩選。以工作日期、建立時<br/>間、ID 依序降冪分頁；nextCursor 為空表示最後<br/>一頁。游標綁定使用者與篩選條件。"]
  h --- n2["操作：GET /api/v1/journals/{journalId}<br/>用途：以日誌 ID 開啟原文，查無資料或無閱讀權<br/>均回 404。"]
```

```mermaid
flowchart LR
  h["資源名稱與 API 骨架"]
  h --- n0["此契約的 groupId 指發布用的 LINE 群組上下文<br/>。人員名單記錄當日配置；creatorId 與 <br/>createdAt 由服務端產生。工作列可省略 <br/>reference。原關聯失權後，讀取端省略該列的 <br/>reference。今日與明日清單合計至少一項。完整<br/>欄位見 Journal 契約。"]
  click n0 "../../../api/contract/journal/journals.yaml"
  h --- n1["後端沿用既有群組、儲存庫、專案及 Firestore <br/>設定。日誌明細與建立收據在同一交易保存。跨群<br/>組 Feed 依當下可讀群組分別查詢，再合併排序。<br/>單頁最多 20 篇，成本隨可讀群組數增加。缺少必<br/>要服務設定時回 503。"]
```

建立日誌時，前端透過 `GET /api/v1/groups/{groupId}/journal-reference-targets` 分頁讀取可引用的儲存庫與專案。後端核對群組資格，依序回傳兩種選項的類型、識別碼與名稱。前端依 `nextCursor` 讀到結尾，發布時由後端再次核對選取項目。
