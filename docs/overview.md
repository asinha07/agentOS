AgentOS — Overview

Vision: Docker for AI agents. Standardize packaging (.agent), runtime contract, and a CLI-first developer experience.

Quickstart
- Build/run built-in agents: `agent-go run startup-builder`, `agent-go run research-agent`
- Package an agent: `agent-go build agents/startup-builder`
- Install from artifact: `agent-go install agents/startup-builder/dist/startup-builder-0.1.0.agent`

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

