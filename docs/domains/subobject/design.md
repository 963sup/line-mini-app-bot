# 子工項設計

<a id="單層子物件"></a>

```mermaid
flowchart LR
  h["單層子物件"]
  h --- n0["使用者手動建立工序；不提供 AI 推測、孫層、刪<br/>除、分派或階層遷移。"]
```

<a id="數量與進度"></a>

```mermaid
flowchart LR
  h["數量與進度"]
  h --- n0["子物件只有名稱、施作／非施作類型、數量、單位<br/>及完成數。數量與單位獨立於父物件：例如父「光<br/>纖」總量140 SET，子「鋪設」可用公尺、「熔接<br/>」可用點；範例不代表自動換算，也不要求每個子<br/>總量小於父總量。"]
  h --- n1["父物件完成數維持獨立人工維護，子項只顯示工序<br/>進度；不加總、平均、取最小比例或套權重推算父<br/>完成數。新增或更新子項不改父的數量、完成數、<br/>狀態或版本；父改類型／數量也不重設子進度。父<br/>類型不限制是否能有子項。"]
  h --- n2["數值沿用物件進度規則：精確十進位字串，最多18<br/>位整數及6位小數，不用浮點；總量大於0，完成數<br/>介於0與總量。新施作預設完成0，後端以原生<br/>switch衍生三態；非施作在畫面顯示不適用。完成<br/>數大於0時不能直接改為非施作，須先明確歸零保<br/>存；轉為施作須提供完成數。修改是設定絕對完成<br/>量，不是累加。"]
  click n2 "../object/design.md#進度與歷史相容"
  h --- n3["子項只接受父建立者寫入。交易內仍核對父歸屬與<br/>管理權、容器及群組綁定；更新只檢查並遞增子版<br/>本。"]
```

<a id="api-與保存-1"></a>

```mermaid
flowchart LR
  h["API 與保存"]
  h --- n0["權威來源為子物件契約，同步生成既有Go／<br/>TypeScript client："]
  click n0 "../../../api/contract/subobject/subobjects.yaml"
  h --- n1["GET／POST /api/v1/repositories/{repositoryId<br/>}/objects/{objectId}/subobjects。"]
  h --- n2["GET／PATCH /api/v1/repositories/{<br/>repositoryId}/objects/{objectId}/subobjects/<br/>{id}。"]
  h --- n3["POST只收內容及建立鍵；PATCH收內容、可選完成<br/>數與expectedVersion，不接受獨立status。"]
  h --- n4["資料位於 channels/{channelId}/repositories/{<br/>repositoryId}/objects/{objectId}/subobjects/<br/>{id}；同父的 subobjectCreationReceipts 保存<br/>建立收據，不使用父物件收據或共享編號。子物件<br/>沒有再向下建立子項的入口。舊物件不自動生成或遷移成子層<br/>。"]
  h --- n5["Firestore刪父文件不會自動清子集合；本功能沒<br/>有父刪除或cascade操作。未來清除任務必須另行<br/>盤點並取得子項、收據及變更紀錄的精確授權。"]
```

<a id="畫面與驗收-1"></a>

```mermaid
flowchart LR
  h["畫面與驗收"]
  h --- n0["物件詳情與展開工序使用同層的兩個互動控制，不<br/>在按鈕內嵌按鈕；用既有shadcn公共props組合，<br/>不更動官方primitives。展開後才載入單一父的子<br/>列表，保留完整長名稱、鍵盤操作及載入失敗重試<br/>；不在首頁逐物件查子資料。"]
  h --- n1["子項新增／編修只包含必要業務欄位。待寫入或建<br/>立結果未知時阻止切父或收合卸載，未知POST用原<br/>內容與原鍵重試；身份／儲存庫切換取消舊請求並<br/>隔離回應。父物件建立期間也會阻止切換。"]
  h --- n2["驗收包含精確數量／非法轉換、跨儲存庫／跨父／<br/>非父建立者拒絕、停用容器、父子版本與scope、<br/>同鍵併發及未知重試、防止子更新改父完成數、<br/>lazy載入及身份切換。正式LINE真機與雲端驗收不<br/>由DOM替身、local emulator或build代替。"]
  h --- n3["功能索引 · 文件索引 · 操作模型"]
  n3 --- n3l0["功能索引"]
  click n3l0 "../README.md"
  n3 --- n3l1["文件索引"]
  click n3l1 "../../README.md"
  n3 --- n3l2["操作模型"]
  click n3l2 "../../architecture/contracts/business-operations.md"
```

物件與子物件都回傳 `requiresProgressReset`。編輯既有資料時，畫面依此旗標提示須先儲存零進度才能改為非施作。新輸入的數量與完成數共用精確十進位表單驗證；後端在更新時仍核對最新的持久化進度。

[跨領域歸屬契約](../../architecture/contracts/object-subobject-ownership.md) · [領域入口](README.md)
