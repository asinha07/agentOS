App Team Demo — Build an App with Five Agents

Overview
- This demo shows how five role-based agents collaborate to produce a lightweight plan for a new app:
  - product-manager → PRD (product_spec.md)
  - be-developer → Backend design (backend_design.md)
  - web-developer → Frontend design (frontend_design.md)
  - qa → Test plan (test_plan.md)
  - code-reviewer → Review notes (review.md)

Install
- Homebrew:
  - brew tap asinha07/homebrew-tap
  - brew install agent
- Linux packages: download .deb/.rpm from Releases and install with apt/rpm
- From source: go build -o agent-go ./cmd/agent

Run Agents Individually
- Product Manager (PRD):
  - agent run product-manager --input "Team todo app"
  - Produces product_spec.md
- Backend Developer (design):
  - agent run be-developer --input "Team todo app"
  - Produces backend_design.md
- Web Developer (design):
  - agent run web-developer --input "Team todo app"
  - Produces frontend_design.md
- QA (test plan):
  - agent run qa --input "Team todo app"
  - Produces test_plan.md
- Code Reviewer (review):
  - agent run code-reviewer --input "Team todo app"
  - Produces review.md

Compose: Run the Team Sequentially
- agent compose --agents product-manager,be-developer,web-developer,qa,code-reviewer --input "Team todo app"
- This runs each agent in order and writes all artifacts in the repo root.

Use Your Preferred Model Provider
- Claude: add `--provider anthropic --model claude-3-5-sonnet-latest` and set `ANTHROPIC_API_KEY`.
- Grok: add `--provider xai --model grok-2` and set `XAI_API_KEY`.

What Happens Under the Hood
- Each agent is packaged with a manifest (agent.yaml) and a role-specific prompt (prompt.md).
- Tools and permissions are declared in agent.yaml; these agents use:
  - file_writer to save artifacts
  - file_reader (for dev/qa/reviewer) to read prior outputs (future enhancement)
  - web_search for PM (offline stub ok)
- The runtime enforces permissions and records events in runs/<run-id>/events.jsonl.

Deploying Your Own Team
- Copy one of the agent folders under agents/ and edit agent.yaml + prompt.md.
- Adjust defaults.output to change the destination artifact.
- Add to compose: agent compose --agents <your-agents> --input "Your app idea"

Next Steps
- Wire stronger dependencies: e.g., be-developer reads product_spec.md.
- Add a CI workflow that runs the team on PR (design doc previews).
- Try the registry: `agent publish product-manager`, then run via `--registry`.
 - See [Model Providers](./models.md) for provider setup and verification.
