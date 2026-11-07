# Cloud Run delivery

Target: containerized Go HTTP service on Cloud Run, images in Artifact Registry,
secrets from Secret Manager and structured logs in Cloud Logging.
The service must listen on the provided PORT and support graceful shutdown.
Use Cloud Run service identity and minimal Google IAM permissions.

Build locally with the existing container-build task once Dockerfile exists.
Deployment needs an explicitly chosen project, region, image and authorization;
this document does not perform provisioning or authorize deployment.
Validate service authentication separately from LINE end-user identity.
