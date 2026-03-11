AgentOS — Overview

Vision: Docker for AI agents. Standardize packaging (.agent), runtime contract, and a CLI-first developer experience.

Quickstart
- Run built-in team agents: `agent run product-manager --input "Team todo app"`
- Compose the team: `agent compose --agents product-manager,be-developer,web-developer,qa,code-reviewer --input "Team todo app"`
- Package an agent: `agent build agents/product-manager`

Architecture
- CLI (`cmd/agent`): init, run, build, install, inspect, logs, publish, compose
- Runtime (cmd/agent for now): loads package, enforces basic permissions, executes tools + mock model
- Tools (`pkg/tools`): plugin interface with built-ins (web_search, file_reader, http_client)
- Memory: per-run JSONL log + KV in `runs/<run-id>/`
- Model adapters: mock adapter (offline); OpenAI stub planned
- Registry: local FS-backed registry with HTTP server skeleton (`registry/server`)
- Workflow: simple engine skeleton in `pkg/workflow`

Package Format
- `.agent` is a tar.gz of agent directory (excluding `dist/`)
- Manifest (`agent.yaml`) uses JSON content for zero-deps parsing in prototype
