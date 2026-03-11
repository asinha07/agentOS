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

## Quick Start

Run the viral demo:

```
agent run startup-builder
```

Use a registry:

```
agent run startup-builder --registry http://localhost:8080
```

## Docs

- [Developer Guide](./developer-guide.md)
- [Architecture](./architecture.md)
- [Specification](./spec.md)
- [Roadmap](./roadmap.md)

