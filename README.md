AgentOS CLI — Portable Agents, One Command Run

[![CI](https://github.com/asinha07/agentOS/actions/workflows/ci.yml/badge.svg)](https://github.com/asinha07/agentOS/actions/workflows/ci.yml)

AgentOS is a CLI-first platform that packages AI agents as portable `.agent` artifacts and runs them consistently across environments. Think “docker for AI agents”: build once, publish to a registry, install and run by name.

Highlights
- Run by name with auto-pull: `./agent-go run <agent> --registry <URL>`
- Portable packaging: build `.agent` artifacts from folders
- Tool, model, and memory abstraction behind a stable runtime contract
- Simple registry integration: search, publish, install

Quick Start
- Build the CLI: `go build -o agent-go ./cmd/agent`
- Run a built-in agent: `./agent-go run startup-builder`
- Auto-install from a registry (if missing):
  - Start the sample registry: `go run registry/server/main.go`
  - `./agent-go run startup-builder --registry http://localhost:8080`
- Use OpenAI (optional): `export OPENAI_API_KEY=...` then run again. The header prints the model in use (e.g., `Model: openai gpt-4.1`).

Core Concepts
- Agent: a packaged unit with manifest (`agent.yaml`), prompts, optional `workflow.yaml`, and assets
- Runtime: executes the agent with permission-aware tool calls, model adapters, and memory
- Tools: capabilities like `web_search`, `file_writer`, `http_client`, `file_reader`
- Model adapters: OpenAI (non-streaming, gpt-4.1 supported), mock fallback for offline
- Registry: minimal HTTP server for discovery and artifact distribution

Commands
- `run <agent|path|artifact>` — run a built-in, directory, or `.agent` artifact
  - Auto-pull: `--registry URL` or `AGENT_REGISTRY` env var for missing agents
  - Input: `--input "your prompt"` (interactive prompt if omitted and workflow asks for input)
  - Model overrides: `--provider openai --model gpt-4.1` (or `anthropic/claude-3-5-sonnet-latest`, `xai/grok-2`)
- `build <agent|--path PATH>` — package a folder into `.agent` under `dist/`
- `install <name|path|artifact>` — install from built-ins, a folder, a `.agent`, or a registry (`--registry URL`)
  - OCI pull: `--oci-ref <ref>` (requires `oras` CLI)
- `inspect <agent|artifact>` — print manifest and basic info
- `logs <agent> [--tail]` — list or tail recent run logs
- `publish <agent>` — copy package to `registry/agents/` (local demo)
  - OCI push: `--oci-ref <ref>` (requires `oras` CLI) and `--sign` to sign blob (requires `cosign` or writes a simple SHA256 .sig)
- `search <query> --registry URL` — search registry for agents
- `init <name>` — scaffold a minimal agent
- `compose --agents a,b [--input X]` — sequentially run multiple agents (demo)

Install/Run From Registry (No Cloning)
- Publisher
  - Package: `./agent-go build agents/my-agent`
  - Publish: `./agent-go publish my-agent && go run registry/server/main.go`
- Consumer
  - Run: `./agent-go run my-agent --registry http://publisher-host:8080`
  - Or install first: `./agent-go install my-agent --registry http://publisher-host:8080`

Agent Package Format
- `.agent` is a tar.gz of an agent folder (excluding `dist/`)
- Typical layout:
  - `agent.yaml` — manifest (JSON content is accepted by the prototype for zero-deps parsing)
  - `prompt.md` — base/system prompt (optional)
  - `workflow.yaml` — JSON content describing steps (optional)
  - `tools/`, `workflows/`, `policies/`, `assets/` — optional directories

Manifest Schema (prototype)
- Minimal fields supported today:
  - `name`, `version`, `description`
  - `entrypoints` (e.g., `{ "run": { "type": "builtin" } }`)
  - `defaults` (e.g., `{ "input": "hello", "path": "README.md", "url": "https://..." }`)
  - `tools`: list of tool names (e.g., `["web_search", "file_writer"]`)
  - `model`: `{ "provider": "openai", "model": "gpt-4.1" }` (mock used if no key)
  - `memory`: `{ "type": "jsonl" | "vector" }` (prototype uses per-run JSONL + KV)
  - `permissions`: `{ "internet": true, "filesystem": "limited" }`

Workflow (optional)
- `workflow.yaml` guides the run. Example JSON content:
  - `{ "steps": [ {"type": "ask_idea"}, {"type": "research", "query": "{idea} competitors"}, {"type": "landing_page", "output": "landing_page.md"} ] }`
- The runtime uses this to shape prompts, tool queries, and artifacts (e.g., landing page path).

Tools (built-ins)
- `web_search` — returns competitor-like results (offline stub)
- `http_client` — HTTP GET with timeout; requires `permissions.internet: true`
- `file_reader` — safe file reads within the working directory; requires FS permission
- `file_writer` — writes files under the working directory; requires FS permission

Model Adapters
- OpenAI (non-streaming): set `OPENAI_API_KEY`. For gpt-4.1, the adapter tries Chat Completions and falls back to Responses API automatically.
- Anthropic Claude (non-streaming): set `ANTHROPIC_API_KEY` (uses Messages API, anthropic-version `2023-06-01`). Example model: `claude-3-5-sonnet-latest`.
- xAI Grok (non-streaming): set `XAI_API_KEY` (uses xAI Chat Completions). Example model: `grok-2` or `grok-beta`.
- Mock: deterministic text for offline demo and graceful fallback

OCI Registry Support (experimental)
- Push: `./agent-go publish <agent> --oci-ref ghcr.io/<org>/<name>:<tag>` (requires `oras` CLI)
- Pull: `./agent-go install <agent> --oci-ref ghcr.io/<org>/<name>:<tag>` (requires `oras` CLI)
- Signing: `--sign` uses `cosign` if available, otherwise writes a `.sig` (SHA256) next to the artifact.

Outputs and Logs
- Each run writes to `runs/<run-id>/`:
  - `events.jsonl` — structured events (`start`, `tool.*`, `model.output`, `final`)
  - `kv.json` — simple key-value snapshot (e.g., topic)
- Tail latest: `./agent-go logs <agent> --tail`

Environment Variables
- `OPENAI_API_KEY` — enable OpenAI model adapter
- `AGENT_REGISTRY` — default registry URL used by `run` and `install`

Build From Source
- Requirements: Go 1.21+
- Build: `go build -o agent-go ./cmd/agent`
- Optional caches for sandboxes: `GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go build -o agent-go ./cmd/agent`

Repository Layout
- `cmd/agent/` — CLI and runtime entrypoint
- `pkg/tools/` — tool plugin interface and built-ins
- `pkg/models/` — model interface + OpenAI adapter (non-streaming)
- `pkg/memory/` — memory interface (adapters placeholders)
- `pkg/workflow/` — workflow spec and engine skeleton
- `pkg/registry/` — registry client
- `registry/server/` — simple HTTP registry (search + artifacts)
- `agents/` — built-in agents for local runs
- `examples/` — example agents and expected outputs
- `docs/` — overview, architecture, spec, and demo docs
  - `docs/developer-guide.md` — step-by-step tutorial to create, package, publish, and run agents

Continuous Integration
- GitHub Actions workflow in `.github/workflows/ci.yml` builds the CLI and vets packages on pushes and PRs.

Releases, SBOM, and Provenance
- Tag a version (e.g., `v0.1.0`) to trigger `.github/workflows/release.yml`.
- GoReleaser builds and publishes archives and Linux packages; Homebrew formula is updated.
- SBOM: Syft generates `sbom.spdx.json` and it is attached to the GitHub Release.
- Provenance: GitHub `actions/attest-build-provenance` produces SLSA v1.0 provenance for `dist/**`.

OCI Artifacts (experimental)
- `.agent` can be pushed as an OCI artifact with annotations, including SHA256.
- Pull verifies the SHA256 from annotations before installation.

Ignore Files
- See `.gitignore` for local runtime state, caches, and artifacts to exclude from source control.

License
- MIT — see `LICENSE`.

Roadmap (selected)
- Replace local cobra shim with upstream dependency and publish binaries
- OCI-based `.agent` artifacts with signatures and provenance (ORAS)
- Stronger sandbox enforcement (FS/network/shell/tool)
- Streaming model adapter support and richer error taxonomy
- Extended workflow engine with more step types and conditionals

Contributing
- Issues and PRs welcome. Please include OS, Go version, and reproduction steps.
- For changes touching the spec, update `docs/spec.md` and add an example agent.

License
- Choose a license that fits your goals (e.g., Apache-2.0 or MIT). Add LICENSE at repository root.

Notes
- A small Python prototype also exists in this repo for fast iteration; the Go CLI is the primary product.
Install via Homebrew (after first release)
- brew tap asinha07/homebrew-tap
- brew install agent

Linux packages (after first release)
- Download .deb or .rpm from the GitHub Release for your version and install with your package manager.
