# agent

[domain documentation](../../../../docs/domains/agent/README.md) · [global index](../../../../docs/domains/README.md#domain-matrix)

Choose the next allowed operation based on task goals and current results.

For candidates missing source page numbers, reread authorized pages or report the need for human confirmation.

[Investigation and termination diagram](../../../../docs/proposals/agent.md#design-diagram) · [Maintenance guidance](AGENTS.md)

[Parent directory guide](../README.md)

Status: read-only evidence investigation is implemented. See the [Go scaffold](application/port/investigate.go).

<a id="模型接線"></a>

## Model wiring

Inject a PageReader verifying identity, resources and source versions through [NewInvestigator](../../bootstrap/agent.go), then call Investigate. Set AGENT_PROVIDER to openai or gemini, AGENT_MODEL to that provider's model and the corresponding API key. Read at most one allowed page per step; only quotations present in the source qualify as candidate evidence. Callers still validate business fields and units.

The [use case](application/investigator.go) obtains candidate quotations or no result through consumer-defined `QuoteModel`. It still controls authorization, step limits and source matching. The [model adapter](adapter/genai/evidence.go) reuses shared text generation for prompts and JSON string/null conversion, keeping provider formats outside core contracts.
