<a id="建置容器映像工作指引"></a>

# Container image build guidance

Inherit [parent guidance](../AGENTS.md); preserve Dockerfile build arguments/CA secret passing.

After changes, run checks from the [script overview](../); see the [local README](./) for entry points.

## Applicable engineering conventions

- **Reproducible builds:** Use locked inputs, explicit build contexts and existing image stages; verify the runtime contract rather than only compilation.
- **Security:** Pass required trust material through the established secret mechanism and keep credentials out of layers, logs and build arguments.
- **Artifact lifecycle:** Identify the built revision/image and cleanup temporary artifacts without publishing or deploying implicitly.
