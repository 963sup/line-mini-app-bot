# 工程項目匯入草稿

## 用途與範圍

處理工程清單匯入，從來源解析、候選驗證和人工審核，到確認後透過 Object 使用案例建立物件。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[object_import.go](object_import.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[匯入草稿規格](../../../../docs/domains/repository/object-imports.md) · [現行實作圖：自動匯入操作契約](../../../../docs/domains/repository/object-imports.md#自動匯入操作契約)

## 閱讀入口

驗證不可信候選與來源引用，保存人工審核快照，確認後沿用 Object 建立使用案例逐行建立。
沒有 AI／PDF 解析或背景工作；重送、權限、限制與驗收依 [匯入草稿規格](../../../../docs/domains/repository/object-imports.md)。

[現行實作圖：自動匯入操作契約](../../../../docs/domains/repository/object-imports.md#自動匯入操作契約) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
