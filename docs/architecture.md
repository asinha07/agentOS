Architecture

Components
- Agent CLI (Go + Cobra shim): entrypoint for init/run/build/install/inspect/logs/publish/compose
- Agent Runtime: loads package, validates, initializes tools/memory/model, executes loop
- Tool System: plugin interface with built-ins (web_search, file_reader, http_client)
- Memory Layer: run-scoped JSONL logs + KV; adapters planned (sqlite/redis/vector)
- Model Adapter: mock adapter; OpenAI stub provided
- Registry: local FS registry and HTTP server skeleton
- Workflow Engine: sequential multi-agent orchestrator skeleton

Execution Loop
1) input → 2) build prompt → 3) model inference → 4) detect/execute tools → 5) update memory → repeat/finish

Security
- Permissions in manifest: filesystem, internet, shell, tool allowlist (future)
- Enforced at tool call sites in prototype

