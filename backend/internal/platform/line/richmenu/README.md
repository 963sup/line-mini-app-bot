<a id="rich-menu-同步"></a>

# Rich Menu synchronization

Provides LINE publication. transport/line/navigation defines content; bootstrap composes it and deployment jobs call Sync. No publication during service startup/webhooks.

<a id="行為與安全邊界"></a>

## Behavior and security boundaries

Validate all images, then identify menus by content hash. Prepare every target image before updating aliases/defaults. Alias updates are not atomic; reruns converge. Identical content causes no recreation/upload/switch. Clean up only managed menus associated with current aliases or explicit migrations, after checking references and migrating users. Abort before writes for old aliases targeting unmanaged menus. Abort on unknown default-menu query errors instead of overwriting.

Callers serialize publication per Messaging API channel. Bootstrap uses Firestore leases and a shared job context for LINE calls; cancellation/timeouts stop further operations. External calls remain outside transaction callbacks.

<a id="驗證與入口"></a>

## Verification and entry points

Run `bash scripts/tooling/tool-exec.sh go -C backend test -race ./internal/platform/line/richmenu ./internal/bootstrap ./internal/architecture` from the root. Tests cover staging failures, alias reruns, zero writes at steady state, failed default queries and migration before deletion. Lease integration needs an isolated loopback Firestore emulator; live LINE acceptance is separate.

[Platform entry](../README.md) · [Automatic deployment configuration](../../../../../cloudbuild.yaml) · [Publication guide](../../../../../docs/operations/rich-menu.md)
