<a id="assets-工作指引"></a>

# assets agent guidance

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [line/richmenu](line/richmenu), the [directory guide](README.md) and [parent guidance](../AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Documentation index](../docs/README.md)

Applies to this directory and its descendants. Follow the [root AGENTS.md](../AGENTS.md), then nearer directory guidance. User and host instructions take precedence; this file grants no additional authorization for external writes.

<a id="職責與邊界"></a>

## Responsibilities and boundaries

line/richmenu/ stores three menu JPEGs and existing operating guidance. Images are packaged in containers for the backend/cmd/richmenu CLI.

When changing images, check dimensions, click regions and purpose. Asset changes do not update LINE menus; remote publication requires explicit authorization.

<a id="驗證"></a>

## Verification

Read line/richmenu/README.md and check image dimensions and Dockerfile asset paths. Absence from Graphify does not establish that images are missing.

Run commands from the repository root unless stated otherwise. Read the [local README](README.md) and [documentation index](../docs/README.md); report actual results and limitations.

## Applicable engineering conventions

- **Naming and formats:** Keep asset names, dimensions, formats and click-region definitions aligned with their existing consumers.
- **Build reproducibility:** Verify source asset inclusion in generated menus and container contexts; do not edit packaged copies as independent sources.
- **Lifecycle:** Distinguish replacing local assets from publishing remote menus, and preserve a recoverable prior version for an authorized publication.
