FAQ — AgentOS

Q: How do I install AgentOS?
A: Homebrew (macOS/Linux): `brew tap asinha07/homebrew-tap && brew install agent`. Linux packages (.deb/.rpm) are in Releases. Or build from source: `go build -o agent-go ./cmd/agent`.

Q: What’s the fastest demo?
A: `agent team --input "Team todo app"` generates PRD, designs, test plan, review notes in minutes.

Q: How do I run a single agent?
A: `agent run product-manager --input "Your idea"`.

Q: How do I use Claude/Grok instead of OpenAI?
A: Set `ANTHROPIC_API_KEY` or `XAI_API_KEY` and run with `--provider anthropic --model claude-3-5-sonnet-latest` or `--provider xai --model grok-2`.

Q: How do I install an agent from GitHub?
A: `agent install github.com/<owner>/<repo>@<tag>` (downloads the first `.agent` asset from that release). Or search: `agent search --github my-agent`.

Q: How do I publish my agent?
A: Add `agent.yaml` and `prompt.md`. Tag a release with a `.agent` file (see [Template Repo](./template-repo.md)). Then others can `agent install github.com/<you>/<repo>@<tag>`.

Q: What is `.agent`?
A: A portable archive of your agent (manifest + prompts + optional assets). It’s the contract between creation and runtime — like a container image for agents.

Q: How does security work?
A: Agents declare permissions in `agent.yaml` (filesystem, network, allowed tools). The runtime enforces them (see [Security](./security.md)).

Q: Where do logs go?
A: `runs/<run-id>/events.jsonl` (structured events) and `kv.json` (simple key-value state).

Q: How do I search and install from a custom registry?
A: Run the demo registry: `go run registry/server/main.go`. Publish with `agent publish <agent>`. Consumers can `agent search --registry <url>` and `agent install <name> --registry <url>`.

