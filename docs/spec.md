Open Agent Specification (Proto)

Manifest (agent.yaml)
- name, version, description
- entrypoints, defaults
- tools: names or objects with {name}
- model: {provider, model}
- memory: {type}
- permissions: {internet: bool, filesystem: 'limited'|'rw'|bool}

Package
- Directory packaged as `.agent` (tar.gz). Includes agent.yaml, prompt.md, tools/, workflows/, policies/.

Runtime Contract (subset)
- Env: run id; outputs in `runs/<run-id>/`
- Events: JSONL log with start/tool/model/final
- Permissions enforced in tool invocations

Tool Interface (proto)
- Execute(input, ctx) -> result
- Schema() -> input/output schema metadata
- Metadata() -> transport, version

Memory Interface (proto)
- Read(key) -> value
- Write(key, value)
- Query(prefix|params)

