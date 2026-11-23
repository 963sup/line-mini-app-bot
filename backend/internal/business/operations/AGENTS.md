<a id="operations-維護指引"></a>

# Operations maintenance guidance

Applies to this directory/descendants; inherit [business guidance](../AGENTS.md). See the [directory guide](./) for scope.

Describe inputs, results, side effects and retries for new operations. Callers supply trusted identities; resource-owning domains authorize.

Verification: documentation checks links and git diff --check. Implementations verify consistent authorization/validation when calling the same operation from different entry points.

When behavior changes, check the [operation and authorization diagram](../../../../docs/proposals/operations.md#design-diagram) and update affected branches and acceptance criteria.

## Applicable engineering conventions

- **Domain, state and authorization:** Keep extraction/model adapters focused on I/O and translate their failures into application errors; the resource owner retains business authorization.
- **Errors and reliability:** Bound document/model input, propagate cancellation and map provider/process failures to stable application errors rather than leaking SDK-specific failures.
- **Testing and privacy:** Verify malformed, oversized, cancelled and unavailable extraction paths; do not retain source documents or add writes implicitly.
