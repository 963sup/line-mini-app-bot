# LINE / Go / Cloud Run

Source scripts/cloud-env.sh in every shell, or use scripts/tool-exec.sh.
Run bash scripts/bootstrap.sh for Go setup, then mise run doctor.
Install optional tooling using mise run install-cloud or install-mcp.

SDK dependencies belong in the future application's go.mod:
- github.com/line/line-bot-sdk-go/v8
- cloud.google.com/go/firestore
net/http, html/template and log/slog require no extra packages.

The repository has no application module yet. mise run dev expects cmd/server
and defaults PORT to 8080. Cloud Run requires listening on :PORT.
mise run container-build requires an application Dockerfile and a Docker daemon.

Use the pinned gcloud with: mise exec gcloud@592.0.0 -- gcloud ...
Install emulator component with: gcloud components install cloud-firestore-emulator --quiet
Then run mise run firestore-emulator. Only the local application process should set
FIRESTORE_EMULATOR_HOST=127.0.0.1:8085 and a local Google Cloud project ID.
Never carry the emulator variable into a production deployment.

Cloud Run, Artifact Registry and Secret Manager require an actual project, region
and authorized identity. Prefer Cloud Run service identity / ADC, not downloaded
service-account keys. No deployment or remote writes are part of environment setup.
Supply LINE secrets securely at runtime; local HMAC signature verification needs the
actual channel secret, not a proxy-only placeholder. Verify webhook signatures on
the original request bytes, and validate MINI App ID tokens server-side for the
expected channel. Avoid logging tokens, secret values or whole webhook payloads.
