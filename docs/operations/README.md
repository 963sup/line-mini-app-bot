<a id="正式部署與維運"></a>

# Production deployment and operations

Covers production deployment, acceptance/recovery. See [business operations](../domains/operations/README.md) for application operations/PDF tools.

| Task                                            | Entry                                                |
| ----------------------------------------------- | ---------------------------------------------------- |
| Local container build/acceptance                | [Container verification](container-build.md)         |
| Build, trigger, IAM/runtime settings            | [Cloud Run deployment](cloud-run.md)                 |
| Runtime variables, secrets/public configuration | [Production configuration](configuration.md)         |
| Release evidence, devices, traffic/rollback     | [Release acceptance/recovery](release-acceptance.md) |
| Rich Menu sync, aliases/retries                 | [Rich Menu publication](rich-menu.md)                |
| Build/Run/post-deployment sync failures         | [Production troubleshooting](troubleshooting.md)     |

Root [cloudbuild.yaml](../../cloudbuild.yaml) supports manual/custom triggers; [Dockerfile](../../Dockerfile) defines containers and [firestore.indexes.json](../../firestore.indexes.json) defines indexes.

Observe target environments for actual settings/results. See [LINE integration](../integrations/line/README.md) for platform wiring and [local troubleshooting](../development/environment/troubleshooting.md) for tools.

[Documentation entry](../README.md) · [Guidance](AGENTS.md)
