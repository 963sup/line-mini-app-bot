<a id="維運文件指引"></a>

# Operations documentation guidance

Follow [documentation guidance](../AGENTS.md); find procedures through the [production index](./).

Runbooks specify targets, operators, prerequisites, success evidence/recovery. Distinguish declarations, live settings and acceptance. Link platform contracts to integrations; preserve authoritative environment names, versions and secret bindings.

Verify command sources, order, links/anchors. Moving documents does not apply live settings.

For root `cloudbuild.yaml`, `Dockerfile` or `firestore.indexes.json`, check [deployment decisions](../decisions/0003-cloud-run-primary.md), target build entry/upload scope. Deploy to Cloud Run with separate build/runtime accounts and Secret Manager bindings.

## Applicable engineering conventions

- **Deployment and versions:** Record the exact artifact/revision, target environment, health evidence and rollback path for an authorized deployment.
- **Migration and cache:** Describe expand-contract ordering, index/data prerequisites and any required cache invalidation before retiring old structures.
- **Security and observability:** Keep least-privilege identities and secret references explicit; use correlated structured logs, traces and bounded metrics for diagnosis.
