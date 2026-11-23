# operations

[domain documentation](../../../../docs/domains/operations/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Provides reusable typed operations delegating to domains owning business rules.

[Operation and authorization diagram](../../../../docs/proposals/operations.md#design-diagram) · [Maintenance guidance](AGENTS.md)

[Parent directory guide](../README.md)

<a id="pdf-文字擷取基礎接線"></a>

## PDF text extraction wiring

`application.DocumentTextExtractor` defines text extraction; `adapter/poppler` implements it using runtime `pdftotext`. Bootstrap injects the adapter into `DocumentTextService` for extraction endpoints. Docker runtime installs `poppler-utils`.

Uses stdin/stdout without temporary files; input/output limits are each 8 MiB with a 30-second timeout. Preserves UTF-8, layout and page breaks. Business JSON conversion and automatic persistence are not implemented.

Poppler output uses `-layout -enc UTF-8 -eol unix`. The adapter validates UTF-8, normalizes CRLF/CR to LF and strips C0/C1 controls except LF, Tab and FF. It neither collapses whitespace nor normalizes NFC. Empty text is valid and does not establish an OCR requirement.

Use `errors.Is` to distinguish missing tools, oversized output, PDF open failure (exit 1), output open failure (2), PDF permissions (3), other conversion errors (including 99), invalid UTF-8 and context cancellation/timeouts; retain application document/provider classifications. Options/exit codes follow official Poppler `utils/pdftotext.1`, checked against master through anonymous Context7. Local version is 25.03.0; Debian Bookworm supplies image packages, whose deployed versions require image acceptance.

<a id="pdf-工具測試"></a>

## PDF tool testing

Sidebar Tools → PDF text extraction opens `/miniapp/poppler`. Authenticated users upload PDFs up to 8 MiB and obtain normalized text through `POST /api/v1/tools/poppler/extractions`. The UI marks page breaks and downloads original text preserving FF; empty text shows a notice. Requests do not store documents/text, call OpenAI/Firestore or have persistent side effects, and may be retried.

HTTP adapters reuse authentication, cookie CSRF and same-origin protection, limiting JSON bodies to 12 MiB. The application validates trusted actors; Poppler converts. Return safe error codes/messages without stderr or document contents.
