# Fly.io deployment from main

Push to `main` -> Repository checks succeed -> Fly production deploy updates the
same Fly App. No branch or preview app is created. The deployment checks out the
exact commit tested by Repository checks, not a potentially newer main commit.
The workflow must be present on the repository's default branch; use main as the
default branch. Pull request checks do not trigger deployment.

## One-time setup

1. Add the Go application (`go.mod`, source, and a production Dockerfile).
   Listen on `0.0.0.0:8080` and handle graceful shutdown. This repository currently
   contains environment configuration, not a deployable LINE application.
2. Authenticate with `fly auth login`, then create an app once with
   `fly apps create <unique-app-name>` (or reuse an existing app).
   Keep the committed fly.toml; do not run fly launch on every deployment.
3. Create an app-scoped deployment token:
   `fly tokens create deploy --app <unique-app-name>`.
4. In GitHub Settings -> Secrets and variables -> Actions, add the repository
   variable `FLY_APP_NAME` and repository secret `FLY_API_TOKEN`.
   Store the token only as a secret and rotate it before expiry.
5. Set `LINE_CHANNEL_SECRET` and `LINE_CHANNEL_ACCESS_TOKEN` in Fly Secrets,
   matching the environment names used by the future application. If using
   Firestore, also configure Google authentication and its project separately;
   Fly does not automatically inherit a Cloud Run service identity.
6. Commit the configuration to main. After the application and prerequisites
   exist, a successful main check run triggers deployment automatically.
7. Configure LINE's webhook as `https://<unique-app-name>.fly.dev/webhook`, using
   the actual route exposed by the application, and verify it in LINE Developers.

`fly.toml` defaults to Tokyo (`nrt`) with one Machine kept running in the primary
region, incurring running compute charges. Choose a region near the database.
One running Machine is not a high-availability guarantee. The app name comes from
`--app`; manual deployments must also supply it. The existing configuration does
not provision the app, credentials, database, or LINE handler.

The workflow fails with an explicit prerequisite error if the application,
Dockerfile or GitHub settings are missing. Syntax checks alone do not establish
that a deployment works. This setup does not change Codex's working branches.

Reference: https://docs.fly.io/launch/continuous-deployment-with-github-actions
