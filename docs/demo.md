AgentOS Killer Demo — startup-builder

Run
- `agent-go run startup-builder` (or `agent run startup-builder` with Python prototype)
- When prompted, enter an idea (e.g., "AI meal planner").

What happens
- Asks for a startup idea (interactive prompt).
- Performs market research via `web_search` (stubbed offline in prototype).
- Generates a product concept using the model adapter (OpenAI if configured, otherwise mock).
- Proposes an MVP architecture (Next.js + FastAPI + OpenAI API + Postgres).
- Writes a landing page to `landing_page.md` using `file_writer`.
- Optionally fetches a URL via `http_client` when `defaults.url` is set and `permissions.internet` is true.

Why this is powerful
- Demonstrates agent packaging (`.agent`), runtime, tool system, memory/events, and a simple workflow.
- Zero code changes; a developer runs one command and sees tangible, multi‑step output.

Agent config (examples/startup-builder/agent.yaml)
- Tools: `web_search`, `file_writer`, `http_client`
- Model: `openai gpt-4o` (falls back to mock if `OPENAI_API_KEY` is not set)
- Memory: `vector` (prototype uses per‑run JSONL + KV)
- Permissions: `internet: true`, `filesystem: limited`

Other commands
- Inspect: `./agent-go inspect startup-builder`
- Install: `./agent-go install startup-builder` (copies built‑in to installed_agents)
- Publish: `./agent-go publish startup-builder` (writes to `registry/agents/`)
- Search (registry server required): `./agent-go search meal --registry http://localhost:8080`
- Install from registry: `./agent-go install startup-builder --registry http://localhost:8080`

Additional demo agents
- research-agent: quick research summary
- coding-agent: reads a local file and summarizes
