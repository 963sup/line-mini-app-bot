# Basic attendance 設計

**Implemented** · [Design](strategy.md) · [Core flow](#decision-flow) · [Representative tests](../../../backend/internal/business/attendance/adapter/firestore/attendance_integration_test.go) · [Module guidance](../../../backend/internal/business/attendance/AGENTS.md)

[Attendance, business features and LINE journeys](journey.md)

<a id="用途與範圍"></a>

## Purpose and scope

Specifications for attendance, punches, corrections, leave and review.

Design entries: [subdomains/investment assumptions](strategy.md) · [context mapping/integration contracts](context-map.md) · [tactical correction design](correction-design.md). This page owns attendance language/boundaries.

<a id="限界上下文"></a>

## Bounded context

Attendance answers who attended in which group/time, what evidence supports it and how much worktime is confirmed. Maintain separate states per person/group. Multiple daily shifts are allowed; monthly reports intersect shifts with report months.

| Boundary           | Facts exchanged                                              | Responsibility                                                                                                                                                                                |
| ------------------ | ------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Identity/groups    | Verified actor and operation-bound group                     | Entry/identity modules authenticate; Attendance determines punch/query/review eligibility and reads reviewer group authorization itself. Membership alone grants no correction-review rights. |
| LINE/MINI App      | Group actions, preflight, personal submission/results        | Entry points translate requests into Attendance operations. Locations evidence individual punches; opening pages does not punch.                                                              |
| Scheduling/payroll | Confirmed net worktime, estimated totals and pending reasons | Attendance currently defines neither expected schedules, overtime eligibility nor pay amounts.                                                                                                |

<a id="通用語言"></a>

## Ubiquitous language

| Term                            | Meaning here                                                                                                           |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Punch Event                     | Identified attempt with action, time/result. Rejected attempts also have results; only accepted actions change shifts. |
| Shift (worktime.Session)        | One person's clock-in/out within one group. Open shifts lack end times; this Session differs from login sessions.      |
| Interval                        | Start/end values for clipping reports, merging overlaps and calculating breaks; value-comparable without identity.     |
| Attendance Record               | Identifiable shift with times, event sequence and correction version; target of corrections.                           |
| Break confirmation              | Personal confirmation that breaks are complete. Empty lists distinguish confirmed no-break from unconfirmed.           |
| Confirmed net / Estimated total | Confirmed time requires complete verifiable evidence. Estimated total already includes confirmed time; never add them. |
| Correction                      | Reasoned changes to a specified record. Request, review decision and effective record are distinct.                    |
| Leave registration              | Personal group leave type, period/reason. Registration neither approves leave nor changes attendance.                  |

<a id="子域與業務能力"></a>

## Subdomains and capabilities

Capabilities below share attendance language/rules by current responsibility. Independent deployments/contexts are not presently needed.

| Capability                    | Decisions owned                                                        | Specification                                          |
| ----------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------ |
| Punches/shifts                | Clock-in, break-start/end and clock-out eligibility; processed retries | [Worktime/transitions](worktime.md)                    |
| Worktime                      | Rebuild shifts from events, clip months and confirm minutes            | [Calculation](worktime.md)                             |
| Corrections/review            | Submission, withdrawal, decisions and approval versions/evidence       | [Evidence](worktime.md), [queries/retries](queries.md) |
| Leave                         | Input/retry validity and personal registration persistence             | [Leave](leave.md)                                      |
| Personal queries/group boards | Read scope and distinct-person working/resting/finished counts         | [Queries](queries.md), [boards](board.md)              |

<a id="一致性與驗收"></a>

## Consistency and acceptance

- **Punches.** Bound state by person/group. needs_position preflight must not create shifts or consume successful receipts. Submit transitions transactionally; redelivery returns original results without extra shifts.
- **Corrections.** Applicants may withdraw. Reviewers require group eligibility and must differ from applicants. Approval transactionally rechecks records, event sequences and overlaps before saving evidence/shifts. Cover self-review, repeat approval and records changed after submission.
- **Reports.** Empty unconfirmed breaks must not count as confirmed net time. Damaged approval evidence yields pending confirmation. Reports/boards are read results, not direct shift-write entry points.

<a id="分層現況與下一個設計決策"></a>

## Current layers and next design decisions

Domains separate [entity](../../../backend/internal/business/attendance/domain/entity/), [valueobject](../../../backend/internal/business/attendance/domain/valueobject/), [service](../../../backend/internal/business/attendance/domain/service/), [event](../../../backend/internal/business/attendance/domain/event/) and [policy](../../../backend/internal/business/attendance/domain/policy/). Records/shifts own identity/state; intervals/locations validate values; calculations, corrections/evidence are domain operations; policies decide reviewers/retry conflicts. [Firestore transactions](../../../backend/internal/business/attendance/adapter/firestore/attendance_correction.go) atomically store decisions/evidence.

The [review use case](../../../backend/internal/business/attendance/application/command/review.go) checks authorization, retries and correction results through transaction ports. HTTP/LINE share the [monthly use case](../../../backend/internal/business/attendance/application/query/monthly.go). Scheduling, overtime/payroll need separately confirmed rules/ownership if adopted.

<a id="相關規格"></a>

## Related specifications

[Local specifications and guidance](../../../api/contract/attendance/AGENTS.md)

<a id="基礎出勤"></a>

## Basic attendance

- Initial group location clock-in/out and personal records by group in direct chat.Commands/postbacks create short-lived requests. MINI App resolves results first,obtains location only for new actions, then submits attendance.Deployment and live LINE/Firestore acceptance are separate.Punch, monthly records, pagination and corrections have implementations;cloud settings, indexes and channel acceptance require verification.Query/retry and worktime/evidence guides own details.Infrastructure guidance records build order, not absence of features.This refactor reuses existing checks without adding test infrastructure.

  - [Queries, web pagination and correction retries](queries.md)

  - [Worktime and correction evidence](worktime.md)

  - [Infrastructure and integration](../../development/environment/verification.md)

<a id="最小操作"></a>

## Minimal operations

- Clock in: backend verifies identity, saves start time and returns result.

- Clock out: backend verifies identity, saves the matching shift end and returns result.

- Personal records: verified actor's own attendance only,including start/end and explicit unfinished status.

- Use trusted backend identity, never client-selected user IDs for authorization.Backend generates punch times. Failed persistence cannot report success.Repeated clicks/redelivery cannot silently duplicate records.Keep open shifts unfinished; never guess clock-out times.

- [Group 出勤 command shows counts; 補卡 opens manual review;漏下班 checks personal open shifts. See board guidance.](board.md)

- [Minimal group leave registration neither approves leave nor changes attendance.](leave.md)

<a id="原始首版範圍與後續實作"></a>

## Original scope and later implementation

Original release excluded schedules, leave, overtime review, payroll,geofencing, face recognition, correction review, admin reports and notifications.Existing-shift corrections, missing-shift entries, manual decisions and withdrawalare now implemented for ended shifts; effective records stay unchanged before approval.Full role management is absent; use group reviewer allowlists.Do not prebuild speculative modules/frameworks; LINE capability lists are references.

<a id="已採用的首版規則"></a>

## Adopted initial rules

- Messaging API: group 打卡 command returns Flex actions with short-lived launch capabilities.MINI App verifies LINE identity/membership and binds a request to person/action.No punch has occurred yet.

- MINI App submits LINE ID token, requestId, latitude, longitude,accuracyMeters and capturedAt to POST /api/v1/attendance/punches after consent.Backend checks identity, ownership/location and executes a Firestore transaction.Denied location creates no record; no radius or physical-presence validation.

- Direct-chat 出勤紀錄 lists personal groups;出勤紀錄 編號 returns that group's latest five.Group queries guide users to direct chat; direct-chat punches are unavailable.

- Calculate each group independently: one open shift per person/group,multiple daily shifts allowed. Cross-day clock-out closes the existing open record.

- Display Asia/Taipei, persist UTC. Punches use server processing time;event time only rejects operations older than the latest processed event.

- Repeat clock-in returns original time; clock-out without an open record rejects.Redelivery reads the original transaction result.

- No personnel allowlist or friendship check. Shared cards require live servermembership verification and reject nonmembers. Show results in MINI App;never push exact coordinates to groups.

<a id="啟用與資料"></a>

## Activation and data

- Initialize attendance with GOOGLE_CLOUD_PROJECT and LINE_MESSAGING_API_CHANNEL_ID;missing either disables attendance with an unavailable reply.FIRESTORE_DATABASE_ID defaults to (default). Reuse Messaging API secret/token and ADC;runtime accounts need database read/write rights.Location punching requires MINI App Channel ID, LIFF ID and openid.Channels must share a Provider for matching user IDs.Verify ID tokens server-side; never trust client userId/groupId.

- Enable Allow bot to join group chats and invite the bot.Members need not friend the bot first; personal records may require direct-chat friendship.ATTENDANCE_MEMBERS_JSON is retired. Direct-chat punches guide users to groups,without a separate personal punching workflow.

- Data under channels/{channelID}/users/{SHA256(userID)}:

- attendance/{SHA256([groupID,eventID])}: schemaVersion,groupID, clockInAt and clockOutAt, omitted for open shifts.

- attendanceStates/{SHA256(groupID)}：recordID、lastEventTime。

- attendanceEvents/{SHA256([groupID,eventID,action])}:code and record snapshot of the original result.

- Location punches use schemaVersion 2 and clockInPosition/clockOutPosition(coordinates, accuracy, sample time); old records remain readable.channels/{channelID}/punchRequests/{requestID} stores original groupID, hashed userKey,eventID, action, eventTime and expiresAt. Redelivery never extends expiry.Retries read original receipts without duplicate punches; expired unfinished requests cannot write.

- [Hash JSON-array composite IDs. Results, records and state commit in one transaction;send replies outside it. Queries use personal paths, group projections, month ranges/cursors.See attendance queries/corrections for contracts; message pagination is not database pagination.Store no reply tokens, full events or names. Locations remain with punches/deduplication,not tracking history or ordinary logs. Retention requires operations decisions;keep deduplication through possible redelivery windows.Optional expiresAt TTL is not immediate expiry; code still checks deadlines.](queries.md)

- Official LINE source: line/line-developers-docs-source main090500a12f9777995006d9fb5cafab8d83a792b2.Checked LIFF opening/query forwarding, server identity and same-Provider user IDs.Device location, Provider settings and live Firestore transactions require real acceptance.

  - [LIFF opening and query forwarding](https://developers.line.biz/en/docs/liff/opening-liff-app/)

  - [Server identity verification](https://developers.line.biz/en/docs/liff/using-user-profile/)

  - [Same-Provider user IDs](https://developers.line.biz/en/docs/messaging-api/getting-user-ids/)

- Enable webhook redelivery in LINE Developers Console.Write/reply failures return 503 for retries; token expiry/single-use limits remain.Deduplication prevents repeated punches, not guaranteed delivery.Check personal records in direct chat for uncertain replies.

<a id="後續產品變更"></a>

## Later product changes

- Initial entry: connect Bot or MINI App first; simultaneous development is not assumed.

- Identity: map LINE users to eligible personnel and define first-use eligibility.

- Time: display zone, cross-day shift matching and multiple daily shifts.

- Results: explicit rejection or replay rules for repeated/inapplicable punches.

- Initial rules are settled above; update this page for later changes.Confirm rules before fixing API, Firestore paths/fields, uniqueness/atomic updates.Do not build every feature before infrastructure stabilizes.

<a id="群組直達-mini-app待上線驗收"></a>

## Direct group MINI App links: production acceptance pending

- Group 打卡 cards open LIFF for clock-in/out, break-start/end and confirmation.Each has a random ten-minute launch token storing group/action, never trusted user ID.Verify identity and live Messaging API membership before exchanging a fixed per-card/user request ID.Reopening never extends expiry. Old ?request= links/postback cards are disabled;request a new card. Existing authorization, five-minute location freshness,state machine and transaction receipts remain.

- Verify MINI App/LINE Login and Messaging API channels share a Provider.Test live membership success, nonmembers, bot departure and API failures.Forwarded cards reject nonmembers. Close LIFF only after confirmed success;permission/network failures keep retry UI. Closing sends no group success message.

<a id="定位前預檢狀態與時序"></a>

## Location preflight, state and sequence

Backend worktime.ApplyEvent owns transitions, never cached frontend state.POST /api/v1/attendance/punches/preflight takes a personally bound requestIdand retains identity, CSRF, expiry and transaction checks.Resolution may persist rejection receipts but never successful punches or locations.

- Existing receipt; any action: replay original result despite changed state; no location.

- No open shift; clock in: needs_position.

- Working/resting; clock in: already_in and original start; no location.

- No open shift; clock out/break/confirm: not_in; no location.

- Working; clock out/start break/confirm breaks: needs_position.

- Working without break start; end break: not_resting; no location.

- Resting; start break/confirm: already_resting/break_incomplete; no location.

- Resting; clock out/end break: needs_position, subject to event ordering.

- Request predates last event; any unreceipted action: stale_event; no location.

- Expired unreceipted or non-owned request; any action: 403; no location/punch.

- Unknown backend state; any action: retryable error; no location.

- Preflight/submit share transactions and state machine; preserve code/message.accepted identifies current/original successful punches.needs_position reserves nothing, creates no event/receipt/shift and guarantees no later success.Persisted rejections replay forever for requestId; changed state needs a new group card.All successful actions, including breaks/confirmation, retain location evidence requirements.

- Current implementation

```mermaid
sequenceDiagram
  participant U as User
  participant W as MINI App
  participant A as Backend
  participant D as Firestore
  participant G as Browser location
  U->>W: Open group action link
  W->>A: Verify identity/membership, exchange fixed requestId
  W->>A: preflight(requestId)
  A->>D: Transaction reads request, receipt, group state and shift
  alt Existing receipt or rejected action
    D-->>A: Original result or persisted rejection
    A-->>W: Definite result without location
  else State change needed
    A-->>W: needs_position without successful write
    W->>G: Obtain current location
    G-->>W: Location or error
    W->>A: Submit same requestId with successful location
    A->>D: Transaction rechecks request, receipt, expiry and current state
    D-->>A: Original result, rejection or atomic success
    A-->>W: Final result
  end
```

- UI: idle → exchanging → checking → locating → submitting → idle.Definite preflight results return checking to idle; failures retain messages.Only needs_position starts location; denial, disconnection or invalid responses stop.If another device acts during location, final submit uses current state/receipt,never duplicating from the preflight snapshot.

- Location failure: no submit; retry same requestId through preflight first.

- Submit disconnect/timeout: unknown result; retry resolves receipt first.Display existing success without locating again.

- Leave/logout: cancel waiting and ignore late HTTP/location results by generation;already submitted transactions cannot be recalled.

- Same-page retries keep requestId and preflight.Navigation, session reset or full reopening clears page caches and revalidates launch/membership;a valid same card still exchanges the same requestId.

- After expiry, owners may replay existing receipts but unfinished requests cannot write.Location acquisition never extends deadlines.

- Preflight/submit: backend 5s, frontend 10s.<br>Launch exchange: backend 8s, frontend 10s.<br>Location: device 10s, frontend guard 15s; do not hide flow defects by waiting longer.

- [Checked official LINE main dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54server user-data guidance; retain server identity verification.React 19.3.1 async handling follows Context7 /reactjs/react.dev useEffect cleanup;Next.js 16.4.5 follows installed use-client guidance.Graphify/Repomix corroborate sources, not device/database acceptance.](https://developers.line.biz/en/docs/liff/using-user-profile/)

<a id="程式入口與命名"></a>

## Code entry points and naming

- PunchResolver.ResolvePunchAttempt handles /punches/preflight.It reads receipts or persists rejection, so is not a pure query.PunchSubmitter.SubmitPunch handles located submissions;both share executePunch transactions and ApplyEvent without duplicate state machines.

- Private punchCommand explicitly carries user, group, event, ReceivedAt,EventTimeMillis ordering, position and request ID. Mode distinguishes resolvePunch/recordPunch,never inferred from absent location. GroupHistoryReader independently serves personal history.RecordsOverlappingMonth includes cross-month/open shifts.verifyAndRebuildRecords checks event/correction evidence and rebuilds without writes;event replay uses worktime.RebuildSessions.

<a id="decision-flow"></a>

<a id="群組打卡的判定流程"></a>

## Group punch decision flow

```mermaid
flowchart TD
    start["Group punch"] --> gate{"Current shift accepts this action?"}
    gate -->|Yes| ok["Record shift event"]
    gate -->|No| stop["Report state conflict"]
```

[Entry adapter](../../../backend/internal/business/attendance/adapter/http/routes.go) · [Decision rules](../../../backend/internal/business/attendance/adapter/firestore/attendance_integration_test.go) · [Application output port](../../../backend/internal/business/attendance/application/port/errors.go) · [Related tests](../../../backend/internal/business/attendance/adapter/firestore/attendance_integration_test.go)

[Domain guidance](../AGENTS.md)
