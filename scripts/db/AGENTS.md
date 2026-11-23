<a id="維護-firestore-索引工作指引"></a>

# Firestore index maintenance guidance

Inherit [parent guidance](../AGENTS.md); confirm target projects/index lists before maintenance.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Data migration:** Match indexes/migrations to the actual query and target project; use expand-contract ordering for incompatible persisted changes.
- **Consistency and recovery:** State repeatability, partial-failure behavior and recovery prerequisites before changing data or index lifecycle.
- **Safety and verification:** Use the established emulator/dry-run path where available and require authorized targets before remote mutation.
