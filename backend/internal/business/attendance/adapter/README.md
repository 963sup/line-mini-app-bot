<a id="attendance-適配器層"></a>

# Attendance adapter layer

Adapters call use cases or implement ports; bootstrap composes them.

<a id="目錄與責任"></a>

## Directories and responsibilities

- [http](http/README.md): HTTP requests, authentication wiring and contract mapping.
- [line](line/README.md): LINE events and message presentation.
- [presentation](presentation/group.go): shared HTTP/LINE group-name presentation.
- [firestore](firestore/README.md): document mapping, queries and transactions.
- [messaging](messaging/README.md): reserved queue entry point; unimplemented.
- [scheduler](scheduler/README.md): reserved scheduler entry point; unimplemented.

<a id="相關入口"></a>

## Related entry points

[module guide](../README.md) · [agent guidance](../AGENTS.md)
