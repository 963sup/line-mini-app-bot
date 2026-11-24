# Documentation

本目錄是專案目前有效的工程文件入口。文件依「規範、架構、設計、決策、工作流程、維運與參考資料」分類。

## 建議閱讀順序

1. [架構總覽](architecture/overview.md)
2. [架構邊界](architecture/boundaries.md)
3. [相依關係](architecture/dependencies.md)
4. [開發流程](workflows/development.md)
5. 依變更範圍閱讀對應規範、設計與維運文件

## 規範

- [規範治理](standards/AGENTS.md)
- [工程執行規範](standards/engineering.md)
- [文件規範](standards/documentation.md)
- [架構規範](standards/architecture.md)
- [後端規範](standards/backend.md)
- [前端規範](standards/frontend.md)
- [API 規範](standards/api.md)
- [安全規範](standards/security.md)
- [測試規範](standards/testing.md)
- [LINE 互動標準](standards/line_interactions.md)

## 架構

- [架構總覽](architecture/overview.md)
- [架構邊界](architecture/boundaries.md)
- [相依關係](architecture/dependencies.md)
- [整合架構](architecture/integration.md)
- [LINE 後端責任與驗收](architecture/line.md)
- [部署架構](architecture/deployment.md)

## DDD 與 Hexagonal Architecture

- [DDD 與 Hexagonal Architecture](ddd_hexagonal/README.md)

## 設計

- [業務領域規則](design/domain_rules.md)
- [LINE 互動介面模型（草案）](design/line_interaction_model.md)
- [Explore 跨領域搜尋設計（草案）](design/explore.md)
- [打卡預檢、出勤月報與補正](design/attendance_punches.md)
- [物件範本版本與套用](design/object_templates.md)
- [授權設計](design/authorization.md)
- [資源模型](design/resource_model.md)
- [狀態轉換模型](design/state_transition.md)
- [事件模型](design/event_model.md)
- [資料模型](design/data_model.md)

## 架構決策

- [決策索引](decisions/README.md)
- [0001：決策範例](decisions/0001_example.md)

## 工作流程

- [開發流程](workflows/development.md)
- [重構流程](workflows/refactoring.md)
- [shadcn MCP 元件實作](workflows/shadcn_mcp.md)
- [Next.js MCP 診斷](workflows/nextjs_mcp.md)
- [驗證流程](workflows/verification.md)
- [發布流程](workflows/release.md)

## 維運

- [設定管理](operations/configuration.md)
- [可觀測性](operations/observability.md)
- [疑難排解](operations/troubleshooting.md)

## 參考資料

- [LINE Bot 聊天指令](references/line_commands.md)
- [外部系統](references/external_systems.md)

## 文件狀態

- `docs/`：目前有效、應持續維護的文件
- `docs-old/`：歷史資料，不作為現行行為或規範的權威來源
- 原始碼與設定：實際行為的最終依據；文件與實作不一致時應同步修正文件
