<a id="google-cloud-與-maps-整合"></a>

# Google Cloud and Maps integration

Use the service table to find the relevant workflow, adoption criteria and acceptance checks. Keep existing Firestore/Cloud Run sources; Cloud Storage belongs under storage and other services here.

<a id="服務規格"></a>

## Service specifications

| Service | Project purpose | Adoption status |
|---|---|---|
| [Cloud Run](../../operations/cloud-run.md) | APIs, webhooks/container deployment | Current deployment authority |
| [Firestore](../firestore/connection.md) | Business state/persistence | Current integration authority |
| [Cloud Storage](../storage/cloud-storage.md) | Files, attachments/exports | Candidate |
| [Cloud Tasks](background-jobs.md#cloud-tasks) | HTTP workers for summaries, notifications/documents with retries/rate limits | Summary queue wired; expand as needed |
| [Cloud Scheduler](background-jobs.md#cloud-scheduler) | Scheduled summaries, reminders/checks/cleanup | On demand |
| [Cloud Run Jobs](background-jobs.md#cloud-run-jobs) | Batch imports/exports/backfills/statistics | On demand |
| [Secret Manager](workload-identity.md#secret-manager) | Versioned/rotatable LINE and server secrets | Deployment rules exist; cloud acceptance separate |
| [IAM/service accounts](workload-identity.md#iam) | Build, deployment, runtime/job Google access | Deployment rules exist; cloud acceptance separate |
| [Cloud Logging](observability.md#cloud-logging) | Structured API/webhook/job success/rejection/failure logs | Operations foundation; collection/retention unverified |
| [Cloud Monitoring](observability.md#cloud-monitoring) | Availability, latency, errors/backlog/job failures | Operations foundation; metrics/alerts unverified |
| [Cloud Trace](observability.md#cloud-trace) | HTTP/Firestore/LINE/AI latency/failure tracing | On demand |
| [Cloud Build](build-and-release.md#cloud-build) | Controlled builds, checks, image push/Cloud Run release | Workflow documented; environment acceptance separate |
| [Artifact Registry](build-and-release.md#artifact-registry) | Exact build images for deployment/rollback | Workflow documented; environment acceptance separate |
| [Document AI](document-processing.md#document-ai) | Extract PDF/scanned/table text and candidate fields | On demand |
| [Vertex AI/Google generative AI](document-processing.md#vertex-ai) | Summaries, understanding, classification/retrieval; no autonomous approval | On demand; choose models/endpoints during implementation |
| [BigQuery](reporting.md#bigquery) | Cross-project attendance/progress/trends without online scans | On demand |
| [Looker Studio/Data Studio](reporting.md#looker-studio) | Authorized attendance/progress/trend reports | On demand |
| [Pub/Sub](event-processing.md#pubsub) | Distribute facts to independent notification/statistics/search consumers | When multiple consumers need one event |
| [Eventarc](event-processing.md#eventarc) | Route cloud events such as completed uploads | When cloud-resource event triggers are needed |
| [Workflows](event-processing.md#workflows) | OCR/model/human-confirmation/notification orchestration | When durable multi-step coordination is needed |
| [Google Maps Platform](google-maps.md) | Site locations, markers/directions; enable address/routes as needed | On demand |
| [Cloud Run functions](event-processing.md#cloud-run-functions) | File events, thumbnails, metadata/lightweight integration; reuse Go when simpler | Candidate, not replacement for main service |
| [reCAPTCHA](recaptcha.md) | Automation abuse risk signals for public/high-risk operations | When public abuse risks warrant it |

<a id="共通責任"></a>

## Shared responsibilities

Go owns trusted identity, resource authorization and decisions. Specify wiring, identity, scope, failure strategy, cost and acceptance during implementation. Verify enablement, pricing, quotas and deployment in the target environment.

Maps means Google Maps Platform; verify reporting/reCAPTCHA names through current official sources. Project names do not pin unadopted SDKs, models, prices, regions or plans. Compare new services against Go responsibilities to avoid duplicate authorization/rules.

<a id="sources"></a>

<a id="查證來源與限制"></a>

## Sources and limitations

2026-11-22: consulted official Maps, Storage signed URLs, Tasks/Pub/Sub, Scheduler, Run Jobs/functions, Workflows, Eventarc, IAM, secrets, observability, builds, Artifact Registry, Document AI, generative AI, BigQuery, report credentials and reCAPTCHA assessment docs. Each page links sources. Online docs are unpinned; recheck APIs/SDKs against project versions during wiring.

Native Google Developer Knowledge/Context7 tools were unavailable; official sites were used. Replaced an unreadable historical report overview with official Data Studio credential guidance. No actual cloud resources, prices or quotas queried; no deployment/functional tests run.

[Integration index](../README.md) · [Architecture index](../../architecture/README.md) · [Guidance](../../architecture/AGENTS.md)
