Agent Manifest (agent.yaml)

Minimal fields (prototype)
```
name: string
version: semver
description: string
entrypoints:
  run:
    type: builtin
defaults:
  input: string
  output: string (optional; file_writer uses this path)
tools: [web_search | file_reader | file_writer | http_client]
model:
  provider: openai | anthropic | xai
  model: e.g., gpt-4.1 | claude-3-5-sonnet-latest | grok-2
memory:
  type: jsonl
permissions:
  filesystem: limited | (future: paths)
  internet: true|false
```

Notes
- JSON content in agent.yaml is supported by the prototype for zero-deps parsing.
- output in defaults is honored by file_writer to save artifacts.
- Multiple tools allowed; permissions control their effects.

