# Intended architecture

LINE Messaging API sends webhooks to a Go net/http service on Cloud Run.
LINE MINI App calls HTTPS APIs on the same service. Firestore stores application
data. html/template can render pages; log/slog emits structured logs.

Separate HTTP/authentication boundaries, application behavior and persistence.
Verify webhook signatures before accepting events. Verify MINI App LINE ID tokens
with LINE's supported verification mechanism; never trust client-supplied user IDs.
Use Secret Manager for server secrets and Cloud Run service identity for Google
access. Prefer narrowly scoped permissions and explicit ownership checks.

This is the intended design. Consult current source and go.mod for implemented
behavior; the environment setup alone does not implement this architecture.
Record consequential choices in [decision records](decisions/README.md).
