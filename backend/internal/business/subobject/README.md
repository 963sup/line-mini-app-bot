# backend/internal/business/subobject 目錄導覽

## 用途與範圍

管理工程物件下的子物件，維護它與上層 Object 的歸屬關係。

## 架構與責任邊界

此目錄維護本領域或子能力的行為。HTTP、LINE 與 Firestore 依需要提供適配，授權和狀態規則歸資源所屬領域。

## 閱讀與修改入口

先查閱[subobject.go](domain/entity/subobject.go)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[現行實作圖：物件與子物件歸屬](../../../../docs/domains/repository/objects.md#物件與子物件歸屬)

## 閱讀入口

從 [firestore](firestore)、[http](http)、[subobject.go](domain/entity/subobject.go)、[subobject_test.go](application/subobject_test.go) 開始閱讀。修改前請參考[本目錄工作指引](AGENTS.md)。

[現行實作圖：物件與子物件歸屬](../../../../docs/domains/repository/objects.md#物件與子物件歸屬) · [未來設計圖](../../../../docs/diagrams.md#未來設計)
