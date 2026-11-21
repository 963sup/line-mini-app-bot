# 物件與子物件歸屬契約

<a id="物件與子物件歸屬"></a>

## 物件與子物件歸屬

父物件與子項都是同一個儲存庫內的 `Object`。子項透過 `parentId` 指向父物件，兩者分別維護數量、單位、完成數與版本。

- 使用 `GET /api/v1/repositories/{repositoryId}/objects?parentId={id}` 分頁列出直接子項。
- 建立子項使用 `POST /api/v1/repositories/{repositoryId}/objects`，在請求本文帶入 `parentId`。編輯子項使用其自身的 Object ID 與版本。
- 專案與工作日誌都引用子項自己的 Object ID。
