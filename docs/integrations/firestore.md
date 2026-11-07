# Firestore integration

Manage the Go Firestore SDK in go.mod. Reuse an appropriately scoped client,
propagate contexts and close owned resources. Define document ownership and
validate it server-side: server SDK access requires application authorization.
Use transactions where concurrent updates require atomic read/modify/write;
transaction callbacks may retry, so keep external side effects outside them.

Develop against an explicitly selected local emulator and test project.
Start the existing `firestore-emulator` mise task only after its prerequisites
are installed. Consult current emulator documentation for Java requirements.
Do not treat emulator success as production IAM or index readiness.
