# Fly.io deployment from main

Fly.io is the only application deployment target. Google Cloud is used for
Firestore access, not application hosting. Runtime application secrets are
injected through Fly Secrets; Google authentication must be configured separately.

Follow [the foundation plan](../development/foundation.md): first establish the
HTTP skeleton deployment, then data/LINE connectivity, then attendance behavior.
Configuration files and a deployment workflow are not evidence of provisioned resources.

Push to `main` -> Repository checks succeed -> Fly production deploy updates the
same Fly App. No branch or preview app is created. The deployment checks out the
exact commit tested by Repository checks, not a potentially newer main commit.
The workflow must be present on the repository's default branch; use main as the
default branch. Pull request checks do not trigger deployment.

## One-time setup

1. The repository already contains `go.mod`, a minimal HTTP server and Dockerfile.
   The server listens on port 8080 by default and supports graceful shutdown.
   LINE and Firestore application integration is not implemented yet; deploying
   the HTTP skeleton does not establish that these integrations work.
2. Authenticate with `fly auth login`, then create an app once with
   `fly apps create <unique-app-name>` (or reuse an existing app).
   Keep the committed fly.toml; do not run fly launch on every deployment.
3. Create an app-scoped deployment token:
   `fly tokens create deploy --app <unique-app-name>`.
4. In GitHub Settings -> Secrets and variables -> Actions, add the repository
   variable `FLY_APP_NAME` and repository secret `FLY_API_TOKEN`.
   Store the token only as a secret and rotate it before expiry.
5. Set `LINE_MESSAGING_API_CHANNEL_SECRET` and `LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN` in Fly Secrets,
   matching the environment names used by the future application. If using
   Firestore, also configure Google authentication and its project separately;
   Fly does not automatically provide Google Application Default Credentials.
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

## MCP operations

See [Fly MCP](../tooling/mcp/fly.md) for OAuth setup, account/resource inspection
and transport limitations. MCP connection does not configure the GitHub deployment
secret or replace the CI deployment workflow.
