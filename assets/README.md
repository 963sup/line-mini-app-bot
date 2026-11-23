<a id="應用素材"></a>

# Application assets

<a id="用途與範圍"></a>

## Purpose and scope

Stores assets used by the application and LINE interface.

<a id="架構與責任邊界"></a>

## Architecture and responsibility boundaries

Assets are maintained here; frontend/backend implementations own feature behavior.

<a id="閱讀與修改入口"></a>

## Reading and editing entry points

Start with [line/richmenu](line/richmenu) and [local guidance](AGENTS.md).

<a id="相關規格"></a>

## Related specifications

[Local specifications and guidance](../docs/integrations/line/setup.md)

<a id="目錄導覽與維護"></a>

## Directory guide and maintenance

line/richmenu/ stores three menu JPEGs and existing operating guidance. Images are packaged in containers for the backend/cmd/richmenu CLI.

When changing images, check dimensions, click regions and purpose. Asset changes do not update LINE menus; remote publication requires explicit authorization.

Read line/richmenu/README.md and check image dimensions and Dockerfile asset paths. Absence from Graphify does not establish that images are missing.

[Local guidance](AGENTS.md) · [Root README](../README.md) · [Root guidance](../AGENTS.md)
