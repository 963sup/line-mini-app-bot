<a id="生成-api-程式碼前端資源與文件圖工作指引"></a>

# API code, frontend asset and document graph generation guidance

Inherit [parent guidance](../AGENTS.md); check artifact/consumer paths after generator changes.

After changes, run checks from the [script overview](../README.md); see the [local README](README.md) for entry points.

## Applicable engineering conventions

- **Code generation:** Derive outputs only from declared schemas/assets/sources using pinned tools, then compare artifact paths and consumer references.
- **Determinism and types:** Preserve generated names, serialization and target-language signatures; repeated runs on identical inputs must not introduce unexplained drift.
- **Lifecycle:** Keep temporary output isolated and clean it after failure; never hand-edit generated files as a substitute for fixing their source.
