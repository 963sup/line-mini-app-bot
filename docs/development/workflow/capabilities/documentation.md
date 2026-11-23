# documentation 能力指引

依[任務選擇表](../../../../README.md#依任務選擇能力)選用；共同執行與權限規則由[根指引](../../../../AGENTS.md)維護。

## 執行要求

1. 編輯前讀取適用的 AGENTS.md、[文件規範](../../../standards/documentation.md)與[驗收矩陣](../../tooling/testing/commands.md)。文件規範統一維護編寫、所有權、自主修正及驗證要求，本頁不另訂一份。
2. 僅修改任務範圍內的文件、圖表與必要直接引用，保留既有變更；擴充檔案範圍前確認所有權。需改程式、工具或設定時，先確認其任務授權。
3. 程式與外部技術敘述依根指引查證；Codex 問題另查 [OpenAI 文件指引](../../tooling/agents/openai-docs.md)。業務文件先讀[操作模型](../../../contracts/business-operations.md)及[符合性流程](../implementation.md#business-operation-compliance)，不自行新增產品決策。證據不足或衝突時列出具體待確認事項。
4. 依文件規範驗證差異、連結、錨點、來源及相關命令，確認精簡後保留有效知識、限制與導航。透過 `scripts/tooling/tool-exec.sh` 執行專案工具；交付列出修改位置、依據版本、實際檢查及限制，未執行者不得宣稱通過。
5. 圖表依 [Mermaid 指引](../../tooling/shared/mermaid.md)保留可編輯來源，區分現況、提案與未知；選擇適合問題的圖形，核對權威來源、引用及交易邊界，並本機渲染及檢查語意。
6. 僅在明確授權下依 [Wiki 發布流程](../../tooling/shared/github.md#wiki-publication)編輯 Wiki，回報網址與核實版本；此授權不延伸至儲存庫推送或部署。使用目前聊天的模型、權限與可用工具。
