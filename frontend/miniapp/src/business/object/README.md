# 儲存庫物件

## 用途與範圍

提供工程物件清單、詳情、編輯與 PDF 匯入畫面。

## 架構與責任邊界

在本功能就近維護畫面與互動，路由由 pages 組合，權限與結果以後端 API 為準。

## 閱讀與修改入口

先查閱[Import.tsx](Import.tsx)、[View.tsx](View.tsx)、[pdf-table.ts](pdf-table.ts)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[物件規格](../../../../../docs/domains/object/design.md) · [PDF匯入規格](../../../../../docs/domains/objectimport/design.md)

提供物件清單、詳情、手動新增及建立者編輯；欄位、權限與狀態依 [物件規格](../../../../../docs/domains/object/design.md)。另提供選擇文字型PDF後本機擷取明細、自動解析保存並顯示結果／編修；規則依 [PDF匯入規格](../../../../../docs/domains/objectimport/design.md)。亦支援 XLSX、CSV、TSV、Tab 分欄 TXT 與 Markdown 表格，由後端依欄名解析；掃描 OCR 尚未支援。

`View` 依儲存庫及身分世代重設狀態，未知建立結果保留原請求鍵及內容；父層導覽在建立結果未確認時保持目前表單。沿用既有 typed client、請求取消及官方 UI 元件組合。

[現行實作圖：物件與子物件歸屬](../../../../../docs/architecture/contracts/object-subobject-ownership.md#物件與子物件歸屬) · [未來設計圖](../../../../../docs/diagrams.md#未來設計)

[上層導覽](../README.md)
