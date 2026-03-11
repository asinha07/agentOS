Security & Permissions

Permissions in agent.yaml
- filesystem: allowlist with `limited` (workdir) or explicit paths; ro/rw considered in future.
- internet: boolean egress permission (HTTP tools like `http_client`).
- tools: allowed tool names (future granularity: per-tool scopes).

Runtime enforcement
- File operations go through the runtime’s tool layer; restricted to workdir when `limited`.
- Network operations via permitted tools only (e.g., `http_client`).
- Audit log: permission-relevant events recorded in `runs/<id>/events.jsonl`.

Best practices
- Default to least privilege (filesystem limited, internet false) and opt-in per agent.
- Use environment variables for provider keys and avoid hardcoding secrets in agents.

Roadmap
- OS-level sandboxing (seccomp/AppArmor/macOS sandbox), brokered filesystem/network, shell restrictions.
- org-wide policy presets and deny logs.

