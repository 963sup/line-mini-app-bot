# docs 工作指引

本目錄沿用[根目錄指引](../AGENTS.md)。文件修改遵循[文件規範](standards/documentation.md)；讀者入口見[按任務開始](README.md#按任務開始)。

## 職責與邊界

單一領域的行為與授權由 domains 維護；跨領域契約由 architecture 維護。已採用的重大決策放在 decisions，待評估的能力放在 proposals。功能操作引用[業務操作模型](architecture/contracts/business-operations.md)，並在領域頁描述具體規則與證據。工具版本、連線設定與機密不得在各頁另立副本。

## 主題子目錄

README 引導讀者選取主題，AGENTS 補充上層未涵蓋的修改規則。新增或搬移頁面時更新所屬分類索引及直接引用；入口變更時才更新上層導覽。區分現行實作、待驗收行為與未來設計，歷史變更由 Git 查閱。

## 驗證

核對相對路徑、來源、標題錨點與索引。改圖時依[圖表規範](standards/documentation.md)驗證圖意與渲染；命令與證據依[驗收矩陣](development/acceptance.md#變更驗收矩陣)。
