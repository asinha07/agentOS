---
title: AgentOS
---

# AgentOS CLI

Portable AI agents with a Docker-like developer experience.

## Install

- Homebrew (macOS/Linux):
  - `brew tap asinha07/homebrew-tap`
  - `brew install agent`

- Linux packages:
  - Download `.deb` or `.rpm` from Releases

- From source:
  - `go build -o agent-go ./cmd/agent`

![Team Demo](./assets/demo-team.gif)

## Quick Start

Install and test:

```
brew tap asinha07/homebrew-tap
brew install agent
agent team --input "Team todo app"
```

Run the viral demo:

```
agent run startup-builder
```

Search and install from GitHub:

```
agent search --github product
agent install github.com/owner/repo@v1.0.0
```

Use a registry:

```
agent run product-manager --registry http://localhost:8080 --input "Team todo app"
```

## Model Providers

Pick your preferred model provider via environment variables and flags.

- OpenAI: set `OPENAI_API_KEY` and run with `--provider openai --model gpt-4.1`.
- Anthropic (Claude): set `ANTHROPIC_API_KEY` and run with `--provider anthropic --model claude-3-5-sonnet-latest`.
- xAI (Grok): set `XAI_API_KEY` and run with `--provider xai --model grok-2`.

You can also set provider/model in `agent.yaml` under `model: { provider, model }`. The CLI header prints the active provider and model.

## Docs

- [Developer Guide](./developer-guide.md)
- [Architecture](./architecture.md)
- [Specification](./spec.md)
- [Roadmap](./roadmap.md)
- [Five‑Agent Team Demo](./demo-app-team.md)
- [Model Providers](./models.md)
