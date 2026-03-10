Contributing to AgentOS

Thank you for your interest in contributing! This guide explains how to set up your environment, propose changes, and run tests.

Prerequisites
- Go 1.21+
- Git

Getting Started
1. Fork and clone the repo
2. Build the CLI: `go build -o agent ./cmd/agent`
3. Run an example: `./agent run startup-builder`

Development Workflow
- Create a branch for your change
- Run `go build ./...` and `go vet ./...`
- Update docs (README, docs/*) if behavior changes
- Open a PR with a clear description and testing notes

Releases
- Maintainers use tags (`v*`) to trigger GoReleaser
- SBOM and SLSA provenance are attached automatically

Code of Conduct
- Be respectful and constructive. Report issues via GitHub Issues.

License
- By contributing, you agree that your contributions are licensed under the MIT License.

