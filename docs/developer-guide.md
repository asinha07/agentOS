Developer Guide — Building and Publishing Agents with AgentOS CLI

Prerequisites
- Go 1.21+

Install and Build the CLI
- Clone the repo and build: `go build -o agent-go ./cmd/agent`
- Verify: `./agent-go --help` (shows commands)

1) Create Your First Agent
- Scaffold: `./agent-go init my-agent`
- This creates `installed_agents/my-agent` with `agent.yaml` and `prompt.md`.
- Open `installed_agents/my-agent/agent.yaml` and set fields:
  - name/version/description
  - defaults: `{ "input": "hello" }`
  - tools: e.g., `["file_writer"]`
  - model: `{ "provider": "openai", "model": "gpt-4.1" }` or keep mock by omitting API key
  - permissions: e.g., `{ "filesystem": "limited" }`

2) Run Locally
- `./agent-go run my-agent --input "hello world"`
- Inspect: `./agent-go inspect my-agent`
- Logs: `./agent-go logs my-agent --tail`

3) Package as .agent
- `./agent-go build installed_agents/my-agent`
- Output: `installed_agents/my-agent/dist/my-agent-<version>.agent`

4) Publish to Local Registry
- Start local registry: `go run registry/server/main.go` (http://localhost:8080)
- Publish: `./agent-go publish my-agent` (copies the artifact to `registry/agents/`)
- Search: `./agent-go search my --registry http://localhost:8080`

5) Install and Run Without Cloning
- From another clone/host:
  - Install by name: `./agent-go install my-agent --registry http://localhost:8080`
  - Or auto-pull on run: `./agent-go run my-agent --registry http://localhost:8080`

OCI Registry (optional)
- Push to OCI: `./agent-go publish my-agent --oci-ref ghcr.io/<org>/my-agent:<tag>` (requires `oras` CLI)
- Pull from OCI: `./agent-go install my-agent --oci-ref ghcr.io/<org>/my-agent:<tag>` (requires `oras` CLI)
- Sign artifacts: add `--sign` (uses `cosign` if available, or writes a SHA256 `.sig`)

6) Use OpenAI (Optional)
- Export key in your shell: `export OPENAI_API_KEY=sk-...`
- Run again; header shows `Model: openai gpt-4.1`.
 - Override on run: `--provider anthropic --model claude-3-5-sonnet-latest` (or `--provider xai --model grok-2`)

7) Add a Workflow
- Create `workflow.yaml` in your agent folder (JSON content). Example:
  ```json
  {
    "steps": [
      {"type": "ask_input"},
      {"type": "research", "query": "{idea} competitors"},
      {"type": "landing_page", "output": "landing_page.md"}
    ]
  }
  ```
- The runtime will prompt for input (if omitted), run research with the templated query, and write the landing page to the path you specify.

8) Write or Use Tools
- Built-ins: `web_search` (stub), `http_client`, `file_reader`, `file_writer`.
- Enable permissions in `agent.yaml`:
  - `internet: true` for HTTP
  - `filesystem: "limited"` for local reads/writes inside the working directory

9) Recommended Project Hygiene
- Add `.gitignore` (ignore `runs/`, `installed_agents/`, `**/dist/`, `.gocache/`, `.gomodcache/`, `landing_page.md`, `registry/agents/*.agent`).
- Use CI to build on PRs (see `.github/workflows/ci.yml`).

Troubleshooting
- If OpenAI not used: verify `OPENAI_API_KEY` in the same shell; `./agent-go inspect <agent>` shows the model; reinstall `./agent-go install <agent>` to pick up config changes.
- If auto-pull fails: ensure registry server is running and reachable and that your agent was published to `registry/agents/`.
