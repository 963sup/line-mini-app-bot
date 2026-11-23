<a id="維護-firestore-索引"></a>

# Maintain Firestore indexes

Entry: ensure-attendance-indexes.py.

- [ensure-attendance-indexes.py](ensure-attendance-indexes.py)
- [ensure-field-policies.py](ensure-field-policies.py) applies the manifest's explicit field TTL and index exemptions, then polls until TTL is ACTIVE and indexes use the declared configuration. It fails on repair-required TTL or a 30-minute timeout per field. Wildcard defaults are excluded.

`cloudbuild.yaml` runs field policies before deploying. For an independently authorized application, run `bash scripts/tooling/tool-exec.sh python3 scripts/db/ensure-field-policies.py PROJECT DATABASE` with an OAuth token in `GOOGLE_OAUTH_ACCESS_TOKEN` or a metadata-server identity. The database defaults to `(default)`. This changes cloud configuration and may delete expired documents; local tests do not apply policies or prove cloud readiness. Repeating it preserves policies already ready; a partial failure can be resumed with the same manifest.

[Script overview](../README.md) · [Local guidance](AGENTS.md)
