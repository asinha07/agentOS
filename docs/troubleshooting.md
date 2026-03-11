Troubleshooting

Install issues
- Homebrew tap not found: `brew tap asinha07/homebrew-tap` first; ensure network access.
- Linux package install fails: use the correct arch (amd64/arm64) and distro’s package manager (apt/rpm).

Provider issues
- 401/403 from providers: verify your API key (OPENAI_API_KEY/ANTHROPIC_API_KEY/XAI_API_KEY) and model availability; try a smaller model (e.g., gpt-4.1 → gpt-4.1-mini).
- Wrong model shown: use `--provider/--model` flags or set in `agent.yaml`.

Registry issues
- HTTP registry returns 404: run `go run registry/server/main.go` and republish agents.
- GitHub install rate limits: set `GITHUB_TOKEN` to increase limits.

Build/Release issues
- Release blocked by mod tidy: run `go mod tidy`, commit `go.mod` and `go.sum`. Our CI has a tidy guard.

Logs & Debugging
- Tail events: `agent logs <agent> --tail`.
- Doctor: run `agent doctor` to test provider keys, GitHub API, and registry connectivity.

Still stuck? Open an issue with OS, command, and logs.
