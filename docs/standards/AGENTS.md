# Standards Governance

本指引適用於 `docs/standards/` 及其子目錄，繼承[全域指引](../../AGENTS.md)與[文件指引](../AGENTS.md)。

## Purpose

以最低必要複雜度建立明確、一致、可執行、可驗證的工程約束，在控制風險的同時保留合理的決策與實作自主性。

## Authority

- MUST 為每項規範建立唯一權威來源。
- MUST NOT 重複定義已有權威來源的規範。
- MUST NOT 建立相互矛盾或權責重疊的規範。
- MUST NOT 以局部規範削弱上層強制約束。
- MUST NOT 未經授權修改既有規範的語意、強度或適用範圍。

## Structure

- MUST 依責任領域組織規範，以規範關鍵字表達約束強度。
- MUST NOT 同時以規範領域與規範強度建立平行權威來源。
- MUST NOT 混合工程規範、架構設計、操作流程及參考知識。
- SHOULD 將具有共同目標、適用條件及因果關係的規則集中管理。

## Normative Language

規範關鍵字遵循 RFC 2119 與 RFC 8174。

| 關鍵字     | 規範作用               |
| ---------- | ---------------------- |
| MUST       | 定義必要條件與強制成果 |
| MUST NOT   | 定義不可逾越的邊界     |
| SHOULD     | 定義預設採用的合理方案 |
| SHOULD NOT | 定義原則上應避免的方案 |
| MAY        | 定義明確允許的選擇     |

各關鍵字用於[對應責任領域的規範](../README.md#規範)。

MUST 與 MUST NOT 具有相同約束強度。SHOULD 與 SHOULD NOT 具有相同約束強度。MAY 不授予額外權限，也不覆蓋其他約束。

## Positive and Negative Rules

正向規則（Positive Rules）定義應滿足的條件、責任及成果。

反向規則（Negative Rules）定義禁止行為、不可接受的狀態及應避免的失敗模式。

規範設計 MUST 依實際目標與風險選擇正向、反向或正反搭配，而非追求特定句型或數量比例。

### Rule Selection

- MUST 使用正向規則描述不可缺少的條件、責任、行為及驗收成果。
- MUST 使用反向規則描述不可接受的行為、越界操作及重大失敗模式。
- SHOULD 使用正反搭配處理單一方向不足以消除歧義的情況。
- SHOULD 在正向規則已充分定義必要條件時省略冗餘反向規則。
- SHOULD 在反向規則已充分界定安全邊界時保留其他實作選擇。
- MUST NOT 將正反搭配視為每項規範的固定格式。

### Complementary Pairing

正反搭配應形成互補約束：

- MUST + MUST NOT：必要條件與不可逾越的邊界。
- SHOULD + SHOULD NOT：預設方案與應避免的替代方案。
- MUST + SHOULD NOT：必要成果與應避免的實作方式。
- MUST NOT + SHOULD：禁止路徑與建議替代方案。
- MUST NOT + MAY：強制邊界與明確允許的自主選擇。

每項搭配 MUST 提供獨立且有效的約束資訊。

MUST NOT 使用不同句型重複相同語意。

MUST NOT 使建議性規則或允許性規則削弱強制約束。

### Rule Examples

以下為撰寫示例，不構成額外工程規範。

**架構邊界：**

- MUST preserve bounded-context ownership.
- MUST NOT access another context's internal persistence implementation.

**最小複雜度：**

- SHOULD prefer concrete implementations.
- SHOULD NOT introduce abstractions without demonstrated requirements.

**安全授權：**

- MUST enforce authorization at the backend operation boundary.
- MUST NOT treat frontend visibility as authorization evidence.

**API 契約：**

- MUST define public HTTP contracts in OpenAPI.
- MUST NOT manually modify generated contract artifacts.

**驗證結果：**

- MUST verify affected behavior against acceptance criteria.
- MUST NOT report unexecuted or failed checks as successful.

## Rule Design

建立規範時 MUST 辨識：

1. 核心目標與必要成果。
2. 適用對象與責任邊界。
3. 不可接受的行為與失敗模式。
4. 正向與反向規則是否需要搭配。
5. 約束強度與例外條件。
6. 可採用的驗證方式。

- MUST NOT 使用模糊、矛盾或無法判定適用條件的敘述。
- MUST NOT 將個人偏好、推測性需求或暫時性實作提升為強制規範。
- MUST NOT 為追求形式完整而增加規範。
- MUST NOT 預先列舉所有合法或非法的實作方式。
- SHOULD 優先約束結果與邊界，而非過度指定實作步驟。
- SHOULD 在必要時提供禁止行為的合理替代方案。

## Context Management

- MUST 遵守上層指引的必要閱讀要求。
- SHOULD 依任務範圍載入相關規範。
- SHOULD 優先引用權威來源，而非複製內容。
- MUST NOT 無理由增加 Agent 的上下文負擔。
- MUST NOT 將參考知識與強制規範混為一談。

## Enforcement

- MUST NOT 將規範宣告視為符合規範的證據。
- MUST NOT 以自然語言取代既有可靠的強制檢查。
- SHOULD 優先使用型別系統、編譯器、靜態分析、架構測試、自動化測試及 CI 驗證可機械判定的約束。
- SHOULD 對無法機械判定的規範採用語意審查。
- MUST NOT 為規範驗證引入缺乏實際效益的框架或工具。

## Maintenance

- MUST 在修改前確認權威來源、實際需求、依賴關係及影響範圍。
- MUST 保留規範遷移前的有效語意與約束強度。
- MUST 同步更新受影響的指引、引用及驗證機制。
- MUST 記錄適用例外的範圍、理由、影響與替代方案，並取得必要核准。
- MUST NOT 為合理化不合規實作而修改規範。
- MUST NOT 保留已完成遷移的重複規範或失效引用。

## Verification

規範變更 MUST 驗證：

- 權威來源唯一。
- 正反規則相互補充。
- 不存在語意重複或矛盾。
- 約束強度符合實際風險。
- 適用範圍與例外條件明確。
- 有效規範完整保留。
- 文件引用與驗證機制一致。

執行 `pnpm docs:check` 與 `git diff --check`，並依影響範圍執行相關工程檢查。

MUST NOT 將未執行、失敗或不完整的驗證宣告為通過。

## Decision Principle

以正向規則確立目標，以反向規則控制風險，以最少必要約束保留自主性，以可驗證成果判定合規性。
