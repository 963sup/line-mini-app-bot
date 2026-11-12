# Bot 管理的群組空間契約

適用於本目錄及其子目錄；繼承 [上層指引](../AGENTS.md)，並遵循其引用的根目錄規則。此處只補充本範圍責任；使用者與主機指令優先，不增加外部寫入授權。

## 職責與邊界

- 先讀 [群組 Repository 規格](../../../docs/features/repository.md)；每個群組一個空間，由驗證簽章的 Bot 事件建立與更新。
- 保留既有列表、詳情與分頁契約；不新增人工建立、刪除、換綁、擁有者或角色管理能力。
- Repository ID 與 groupId 分開；互動候選不是完整成員名冊，也不是授權。容器存取資格不自動授予子業務寫入或審核權。
- Member、Label、Milestone 等未來能力先定義操作與擁有者，再設計契約；不得把互動候選直接當作 members API。
- 子業務引用既有 Repository 歸屬，不由 Issue／Discussion 的寫入順便建立容器。
