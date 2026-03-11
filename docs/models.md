Model Providers — OpenAI, Anthropic (Claude), xAI (Grok)

Overview
- AgentOS supports multiple providers behind a single adapter interface. You can fix the provider in agent.yaml or override per-run with flags.

Environment Variables
- OpenAI: `OPENAI_API_KEY`
- Anthropic (Claude): `ANTHROPIC_API_KEY`
- xAI (Grok): `XAI_API_KEY`

Per‑Run Overrides
- OpenAI: `agent run product-manager --provider openai --model gpt-4.1 --input "Team todo app"`
- Claude: `agent run product-manager --provider anthropic --model claude-3-5-sonnet-latest --input "Team todo app"`
- Grok: `agent run product-manager --provider xai --model grok-2 --input "Team todo app"`

Permanent Configuration (agent.yaml)

```
model:
  provider: anthropic
  model: claude-3-5-sonnet-latest
```

Verification
- The run header prints `Model: <provider> <model>`.
- If a provider call fails or a key isn’t set, the runtime falls back to a mock response so demos still run offline.

