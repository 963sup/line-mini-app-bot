<a id="summary-既有入口契約"></a>

# Existing Summary entry contracts

Documents LINE interactions and internal job entry points for group message summaries.

Describes LINE message collection, summarization and internal worker contracts. Internal OpenAPI composes the worker; existing webhooks handle LINE events.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [worker.yaml](worker.yaml) and [local guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Group chat summaries](../../../docs/domains/summary/summary.md) · [Current implementation diagram: data and reliability](../../../docs/domains/summary/summary.md#資料與可靠性)

This document describes existing LINE/internal job entries. It adds neither public OpenAPI operations nor a browser client.
See [group chat summaries](../../../docs/domains/summary/summary.md) for complete product rules and persistence.

| Capability           | Entry and trusted source                                                     | Inputs, rules and results                                                                                                                      | Backend                                                                 |
| -------------------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| MessageCollection    | Group text/edit/unsend events after shared LINE webhook signature validation | Silently collect ordinary text, exclude commands; retain version invalidation for edits/unsends                                                | business/summary/messagecollection, its line/; shared summary/firestore |
| Summarization        | Group summary command through the same webhook                               | Valid group, user, event ID, reply token and freshness checks; deduplicate by event ID, create/schedule jobs; do not reschedule completed jobs | business/summary/summarization/line                                     |
| Summarization worker | POST /internal/summary; Google job identity Bearer authentication            | JSON group (C plus 32 lowercase hex digits), id (64 lowercase hex digits); 2048-byte limit; reject unknown fields and multiple JSON values     | business/summary/summarization/http                                     |

Preserve worker responses: 204 no content; 401 unauthorized; 400 invalid job; 503 summary temporarily unavailable.
Queues send lowercase group/id. Handlers retain Go JSON decoder case matching; refactoring must not change accepted formats.
Jobs retain leases, version rechecks, result persistence and stable LINE retry keys. Existing job retries handle model or delivery failures.
The shared LINE composition entry in business/summary/line dispatches summary commands before message collection. Collection must not depend on summary generation.

There is no existing MINI App screen or public Summary HTTP API; this internal job endpoint is not a browser API.

[Current implementation diagram: data and reliability](../../../docs/domains/summary/summary.md#資料與可靠性)

[Parent directory guide](../README.md)
