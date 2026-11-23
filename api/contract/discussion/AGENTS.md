<a id="discussion-契約指引"></a>

# Discussion contract guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [categories.yaml](categories.yaml), [comments.yaml](comments.yaml), [discussions.yaml](discussions.yaml), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Discussion operation specification](../../../docs/domains/repository/discussion/design.md) · [Current implementation diagram: minimal boundaries and verification](../../../docs/domains/repository/discussion/design.md#最小實作邊界與驗證)

Applies to this directory and descendants. Inherit [parent guidance](../AGENTS.md) and its referenced root rules. This file adds only local responsibilities. User and host instructions take precedence; it grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

- Status: basic topics and single-level comments are wired. Read the [discussion specification](../../../docs/domains/repository/discussion/design.md) first. Categories remain a scaffold; specify other extensions separately.
- Use Discussion for entities and discussions for collections. Topics belong to existing Repositories; this capability defines DiscussionComment reply relationships.
- Decide creation, reads, updates, categorization, replies, answers and locking individually by need. GitHub's full operation set is not a mandatory checklist.
- Specify visibility, edit permissions, reply depth, lifecycle, retries and concurrency. Comments, reactions and marking answers do not constitute formal approval.
- Linking or converting Issues requires separate authorization, identity and traceability semantics; do not directly overwrite or share their states.

[Current implementation diagram: minimal boundaries and verification](../../../docs/domains/repository/discussion/design.md#最小實作邊界與驗證)

## Applicable engineering conventions

- **Contract semantics:** Specify topic/comment lifecycles and edit versus classification permissions independently; never equate answer selection with business approval.
