# 前端規範

MINI App 使用 Next.js Pages Router、React 與靜態輸出。前端負責呈現、互動、即時輸入回饋與必要的頁面狀態；授權與權威業務判斷由後端完成。

## 結構

- `business/` 按業務能力組織頁面與互動。
- `identity/` 擁有登入、Session 與身分切換生命週期。
- `platform/` 封裝 LINE 等平台 SDK。
- `shared/` 只放真正跨功能且不擁有業務規則的能力。
- `shell/` 擁有共用頁框與導航。
- `pages/` 只負責路由組裝。

## 命名

- TypeScript 變數、函式與屬性識別字使用 `camelCase`，例如 `repositoryId`。
- React 元件使用 `PascalCase`，例如 `ServiceSidebar`。
- 外部契約欄位與框架約定的名稱依各自的來源，例如 Next.js 的頁面路徑。

## 資料與狀態

- 使用產生的 API client，不自行維護第二份 DTO 或錯誤碼。
- 身分、資源範圍或查詢條件改變時，取消舊請求並拒收遲到回應。
- 不把可由 props、查詢結果或渲染推導的值另存成 state。
- 寫入操作要區分成功、拒絕、可重試失敗與結果未知；不得因 timeout 自動假設未提交。
- 表單需保留可修正輸入，分清未填、清除與未變更。

## 瀏覽器生命週期

React Strict Mode 下 effect 必須可清理且不重送副作用。離頁、登出、身分世代改變與 BFCache 恢復時，需取消或重新驗證相關工作。定位 watcher、timer、subscription 與 request 都必須有對稱清理。

## UI 與可用性

畫面必須回答使用者目前範圍、操作對象與下一步結果。載入、空資料、失敗、無權限與成功狀態不可混用。鍵盤、焦點、標籤、錯誤提示與非滑鼠操作必須可用；圖表的重要數值不能只靠 hover。

## 驗證

前端變更至少執行 TypeScript 檢查與相關測試。路由、靜態資產或建置設定變更需驗證 production build；LINE 登入、瀏覽器權限與真實裝置行為須另外驗收。

## 前端實作邊界

### Components

- SHOULD reuse existing shadcn components and shared UI primitives.
- MUST NOT modify the source of installed official shadcn/ui components. Implement project-specific behavior and styling at the call site or in a project-owned component using the official component's public API.
- SHOULD prefer accessible, responsive, and consistent interactions.
- SHOULD NOT create custom components when existing components satisfy the requirement.
- MAY introduce reusable UI components when multiple consumers share stable behavior.
- MAY use additional frontend libraries when existing capabilities are insufficient and authorization requirements are satisfied.
- MUST NOT create custom UI primitives when shadcn/ui provides an equivalent component.
- MUST NOT recreate functionality already provided by installed shadcn/ui components.
- MUST NOT duplicate existing components or component variants with equivalent behavior.
- MUST NOT introduce wrapper components without distinct responsibilities.
- MUST NOT introduce alternative UI component libraries without authorization.
- MUST NOT modify shared UI primitives to satisfy feature-specific requirements.

### Styling

- SHOULD NOT introduce independent design tokens for existing semantic roles.
- MUST NOT introduce custom CSS when existing Tailwind CSS utilities satisfy the requirement.
- MUST NOT hardcode colors, spacing, typography, or border values when corresponding shared theme tokens exist.
- MUST NOT introduce parallel design systems or theme definitions.
- MUST NOT duplicate styles already provided by shadcn/ui.
- MUST NOT introduce inline styles for static styling requirements covered by existing utilities.

### State

- SHOULD keep feature-specific state within its owning feature and derive display values from existing state.
- SHOULD NOT introduce global state for feature-local behavior or use Effects for values directly derivable during rendering.
- SHOULD NOT introduce client-side caching or synchronization mechanisms without demonstrated need.
- SHOULD NOT introduce unnecessary client dependencies or state-management frameworks.
- MAY use optimistic UI when operation semantics and recovery behavior support it.
- MUST NOT store values that can be derived directly from existing state, props, or query results.
- MUST NOT duplicate server state across independent client state stores.
- MUST NOT promote feature-local state to global state without a demonstrated cross-feature requirement.
- MUST NOT use Effects solely to synchronize derived state.
- MUST NOT introduce additional state-management libraries without authorization.
- MUST NOT maintain multiple authoritative representations of the same state.
- MUST NOT allow stale responses to overwrite current actor-scoped or resource-scoped state.
- MUST NOT retain private resource data across unauthorized scope changes.

### API and Business Logic

- SHOULD maintain a thin-client architecture and use generated API types.
- SHOULD use explicit feature-level view models where necessary.
- MAY introduce feature-specific view models when generated transport types are insufficient for presentation.
- MUST NOT implement authoritative business rules in the frontend.
- MUST NOT duplicate backend authorization, calculations, or state transitions.
- MUST NOT manually recreate OpenAPI-generated types or API clients.
- MUST NOT bypass the established generated API client.
- MUST NOT introduce redundant API forwarding layers without distinct boundary responsibilities.
- MUST NOT expose persistence-specific models to presentation components.
- MUST NOT maintain independent frontend business models without distinct presentation requirements.

### Abstractions

- MUST NOT introduce abstraction layers without distinct responsibilities.
- MUST NOT create hooks that merely forward existing functions.
- MUST NOT create utility functions that merely rename existing functionality.
- MUST NOT introduce interfaces or types that duplicate existing contracts without semantic differences.
- MUST NOT introduce directories or modules solely to satisfy a reference structure.
- MUST NOT retain unused components, hooks, types, imports, or dependencies.
- MUST NOT introduce speculative extension points without approved requirements.

### Runtime

- MUST NOT introduce per-request server-side execution into the Next.js application.
- MUST NOT introduce Next.js API routes as alternative backend endpoints.
- MUST NOT introduce frontend-owned persistence.
- MUST NOT introduce server-side business logic into the Next.js application.
- MUST NOT introduce runtime dependencies or behavior incompatible with the Pages Router static-export deployment model.

### Minimal Implementation

- MUST NOT introduce new code that duplicates an existing capability without a distinct and necessary responsibility.
- MUST NOT introduce new implementations when existing implementations satisfy the requirements.
- MUST NOT add functionality outside approved requirements.
- MUST NOT increase architectural complexity without demonstrated necessity.
- MUST NOT retain superseded implementations after their consumers have migrated.
- MUST NOT introduce new dependencies when existing dependencies provide equivalent capabilities.
- MUST NOT sacrifice correctness, accessibility, security, or maintainability solely to reduce code size.
- MUST NOT disable required TypeScript strictness or suppress type errors to bypass validation.
