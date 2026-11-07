# LINE integration

Use the LINE SDK v8 messaging_api and webhook packages when the application is
created. Versions belong in go.mod, not mise. Verify request signatures over the
unaltered body using the SDK before handling events; bound body size and handle
redelivery idempotently. Keep channel secrets and access tokens in runtime secrets.

MINI App identity requires server-side LINE ID token verification, including
expected channel/audience and expiry according to current LINE documentation.
Google's idtoken package does not verify LINE ID tokens. Authorize access using
verified identity; never accept client-supplied ownership as proof.
Use Context7/official LINE docs for the exact pinned SDK API. Test bad signatures,
invalid identities, duplicate events and downstream failures.
