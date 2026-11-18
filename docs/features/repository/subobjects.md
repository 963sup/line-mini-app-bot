# 物件的單層子物件

狀態：最小切片已完成本機實作；正式 LINE／手機及雲端整合仍須另行驗收。每個 Object 可展開一層 Subobject，工序由使用者手動建立，不由 AI 推測，不提供孫層、刪除、分派或階層遷移。

## 數量與進度

子物件只有名稱、施作／非施作類型、數量、單位及完成數。數量與單位獨立於父物件：例如父「光纖」總量140 SET，子「鋪設」可用公尺、「熔接」可用點；範例不代表自動換算，也不要求每個子總量小於父總量。

父物件完成數維持獨立人工維護，子項只顯示工序進度；不加總、平均、取最小比例或套權重推算父完成數。新增或更新子項不改父的數量、完成數、狀態或版本；父改類型／數量也不重設子進度。父類型不限制是否能有子項。

數值沿用[物件進度規則](objects.md#進度與歷史相容)：精確十進位字串，最多18位整數及6位小數，不用浮點；總量大於0，完成數介於0與總量。新施作預設完成0，後端以原生switch衍生三態；非施作在畫面顯示不適用。完成數大於0時不能直接改為非施作，須先明確歸零保存；轉為施作須提供完成數。修改是設定絕對完成量，不是累加。

## 操作契約

| 欄位 | List／Get | Create | Update |
| --- | --- | --- | --- |
| Actor／Resource／Action | Session／Bearer可信使用者；指定Repository內父Object的Subobject；讀取 | 同scope建立單層子項 | 同scope更新既有子項 |
| Context／Authority | 當下群組成員、ACTIVE儲存庫、父存在且歸屬同儲存庫 | 同左，且為父物件建立者 | 同左，並核對子expectedVersion |
| Policy／Decision | 沒有當下讀取資格或父不存在不揭露資料；游標不能跨本人／儲存庫／父使用 | 交易內核儲存庫群組綁定、父及作者；client不能指定actor、父歸屬、狀態或版本 | 同交易核父管理權、子版本及數量規則；失敗不改父子資料 |
| Result | 子列表／詳情，沒有寫入副作用 | server產生子ID、creator、version=1、時間與衍生狀態 | 有效內容變更遞增子版本，同值不製造變更；父快照不變 |
| 重送／併發 | 分頁不是總數；首次展開才載入該父子列表 | Idempotency-Key綁本人／repo／父／內容，同鍵同內容返回原子項；同鍵不同內容409，未知結果保留同body/key | expectedVersion過期409，必先重讀再修改；不做盲目覆寫 |
| Event／Audit | 不為讀取建立紀錄 | 建立子項與專用收據同交易提交 | 保存操作者、時間、子版本與變更；不記錄token、不發外部通知 |

父建立者管理所有子項，避免子作者不同而令父無法協調工序；其他目前成員只讀。群組成員資格是當下外部觀察，不能與Firestore組成原子快照；交易內仍核ACTIVE容器、群組綁定及父管理權。前端按鈕不取代後端授權。

## API 與保存

權威來源為[子物件契約](../../../api/contract/subobject/subobjects.yaml)，同步生成既有Go／TypeScript client：

- GET／POST `/api/v1/repositories/{repositoryId}/objects/{objectId}/subobjects`。
- GET／PATCH `/api/v1/repositories/{repositoryId}/objects/{objectId}/subobjects/{id}`。
- POST只收內容及建立鍵；PATCH收內容、可選完成數與expectedVersion，不接受獨立status。

資料位於 `channels/{channelId}/repositories/{repositoryId}/objects/{objectId}/subobjects/{id}`；同父的 `subobjectCreationReceipts` 保存建立收據，不使用父物件收據或共享編號。子物件沒有再向下建立子項的入口。現有Object API、列表與AI匯入不變，舊物件不自動生成或遷移成子層。

Firestore刪父文件不會自動清子集合；本功能沒有父刪除或cascade操作。未來清除任務必須另行盤點並取得子項、收據及變更紀錄的精確授權。

## 畫面與驗收

物件詳情與展開工序使用同層的兩個互動控制，不在按鈕內嵌按鈕；用既有shadcn公共props組合，不更動官方primitives。展開後才載入單一父的子列表，保留完整長名稱、鍵盤操作及載入失敗重試；不在首頁逐物件查子資料。

子項新增／編修只包含必要業務欄位。待寫入或建立結果未知時阻止切父或收合卸載，未知POST用原內容與原鍵重試；身份／儲存庫切換取消舊請求並隔離回應。父既有手動建立與自動匯入阻擋不因新增子項而放寬。

驗收包含精確數量／非法轉換、跨儲存庫／跨父／非父建立者拒絕、停用容器、父子版本與scope、同鍵併發及未知重試、防止子更新改父完成數、lazy載入及身份切換。正式LINE真機與雲端驗收不由DOM替身、local emulator或build代替。

[物件](objects.md) · [功能索引](../README.md) · [操作模型](../../foundation/architecture/business-operation.md)
