<a id="發布驗收工作指引"></a>

# Release acceptance guidance

Inherit [parent guidance](../AGENTS.md); confirm targets before acceptance.

After changes, run checks from the [script overview](../README.md); see the [local README](README.md) for entry points.

## Applicable engineering conventions

- **Versions and compatibility:** Apply the existing compatibility policy and Semantic Versioning where the published artifact uses it; identify the exact tested revision.
- **Delivery workflow:** Keep independently verified milestones and purpose-specific Conventional Commit messages; record review and Definition of Done evidence.
- **Deployment and recovery:** Verify target health and rollback prerequisites for authorized releases, distinguishing a repository push from production acceptance.
