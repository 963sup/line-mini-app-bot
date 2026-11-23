<a id="attendance-領域層"></a>

# Attendance domain layer

Does not reference application, adapters, HTTP or Firestore.

<a id="目錄與責任"></a>

## Directories and responsibilities

- [entity](entity/README.md): business objects, identity, state and invariants.
- [valueobject](valueobject/README.md): identity-free business values and validation.
- [service](service/README.md): cross-object calculation, corrections and evidence checks.
- [event](event/README.md): recorded punch facts.
- [policy](policy/README.md): review eligibility, retries and leave rules.

<a id="相關入口"></a>

## Related entry points

[module guide](../README.md) · [agent guidance](../AGENTS.md)
