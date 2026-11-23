<a id="line-平台能力"></a>

# LINE platform capabilities

Provides shared LINE Group Member API queries and [Rich Menu synchronization](richmenu/README.md). Bootstrap injects existing Membership ports; no transport, identity or business dependencies.

<a id="入口與語意"></a>

## Entry points and semantics

[membership.go](membership.go) queries live membership without treating cached interactions as authorization. Only HTTP 200 with valid JSON and matching userId returns true. 404 returns false; 400, authentication, rate limits, network/invalid-data errors return ErrMembershipUnavailable without authorization. 404 may mean the bot left; it does not prove the user left.

Default timeout is one second; bootstrap sets five for Attendance. Shorter upstream deadlines take precedence. Bootstrap clients reject redirects. ValidGroupID checks ID format only, not membership or permissions.

<a id="責任邊界"></a>

## Responsibility boundaries

transport/line owns signatures, dispatch and reply-token semantics; transport/line/navigation owns shared navigation/Rich Menu definitions. Domain adapters own Flex, Quick Reply and policies. Identity owns MINI App tokens/sessions. Use official SDK message types; no generic message framework was added.

<a id="驗證與來源"></a>

## Verification and sources

Run `bash scripts/tooling/tool-exec.sh go -C backend test ./internal/platform/line ./internal/architecture ./internal/bootstrap ./internal/transport/line/...` from the root. Tests do not call live LINE; real-channel acceptance is separate.

Official LINE [Group Member Profile](https://developers.line.biz/en/reference/messaging-api/#get-group-member-profile) was checked against line-developers-docs-source main revision c15d6d3fa2eff7fcae65554c20be5b1e60f9a10a. SDK remains v8.23.0.

[parent directory guide](../README.md) · [parent guidance](../AGENTS.md)
