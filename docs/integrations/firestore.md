# Firestore integration

Manage the Go Firestore SDK in go.mod. Reuse an appropriately scoped client,
propagate contexts and close owned resources. Define document ownership and
validate it server-side: server SDK access requires application authorization.
Use transactions where concurrent updates require atomic read/modify/write;
transaction callbacks may retry, so keep external side effects outside them.

Configure the selected Google Cloud project, database and application identity
explicitly. Fly.io requires separately supplied Google authentication and
appropriate IAM; do not assume credentials are provided by the deployment host.
