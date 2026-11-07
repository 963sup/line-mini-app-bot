# Cloud Run primary deployment

Cloud Run is the application runtime; see
[ADR 0003](../architecture/decisions/0003-cloud-run-primary.md). The container image
and environment-variable contract are deployment inputs.

## Required setup

1. Select the Google Cloud project, Firestore database location and Cloud Run region.
2. Enable Cloud Run, Artifact Registry, Cloud Build, Firestore and Secret Manager APIs
   as required by the selected delivery path.
3. Create a dedicated runtime service account. Grant only the Firestore permissions
   required by the implemented operations and access to the specific runtime secrets.
4. Store LINE channel credentials in Secret Manager and expose them to the service as
   `LINE_MESSAGING_API_CHANNEL_SECRET` and
   `LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN`. Do not bake values into the image.
5. Build the existing Dockerfile, deploy the image with port `8080`, the runtime
   service account and the selected secret bindings.
6. Record the service URL and configure LINE's webhook as `<service-url>/webhooks/line/<channelID>`.
   Verify the webhook in LINE Developers and inspect Cloud Run logs.

The application currently has no health endpoint. The optional LINE webhook prototype
is not attendance or Firestore integration. Repository configuration does not prove
that a Cloud Run service, IAM binding, database or secret exists.

## Delivery status

The repository does not currently contain a Cloud Run deployment workflow or an
infrastructure definition. Add one only after the project, region, Artifact Registry,
runtime service account and deployment authority are confirmed. Until then, deployment
is a documented manual operation, not an automated capability.

Cloud Run supplies Application Default Credentials through the attached service account;
do not deploy a service-account key file.
