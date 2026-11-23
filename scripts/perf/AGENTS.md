<a id="go-效能剖析工作指引"></a>

# Go profiling guidance

Inherit [parent guidance](../AGENTS.md); profile one Go package and keep output in .artifacts/go.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Measurement:** Profile a defined workload/package and compare before/after results under matching conditions; avoid optimization based only on code shape.
- **Observability:** Use interpretable latency/allocation/CPU evidence and bounded metric dimensions rather than resource IDs as labels.
- **Artifact lifecycle:** Keep profiles in the declared local artifact directory and exclude sensitive payloads or credentials.
