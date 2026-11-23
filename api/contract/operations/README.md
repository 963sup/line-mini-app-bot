<a id="文件工具契約"></a>

# Document tool contracts

`poppler.yaml` defines authenticated PDF text extraction testing. JSON contains a filename and Base64 data; raw PDFs are limited to 8 MiB. Returns normalized text preserving FF page breaks. Does not store documents, invoke AI or create business objects. Cookie sessions require CSRF tokens and same-origin protection.

Sources are composed in `../openapi.yaml`. Follow parent API guidance to generate and verify with `pnpm api:generate` and `pnpm api:check`.
