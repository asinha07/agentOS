Roadmap

## v0.2.0 (Next)

- Security sandbox
  - Enforce filesystem/network/tool policies at runtime
  - Deny logs + human-readable violations
  - Good first issue: add per-tool allowlist checks with tests

- Streaming + UX
  - Streaming model tokens
  - `--verbose` flag for tool/memory event echo
  - Good first issue: implement `--verbose` output gate

- Registry search & UX
  - `agent search` pagination and filters
  - `agent run` auto-install with registry fallback
  - Good first issue: add `--page` and `--limit` to `agent search`

- Conformance & Spec
  - JSON Schemas for agent.yaml & tools
  - `agent test` to validate packages
  - Good first issue: write initial agent.yaml schema and validate in CLI

- OCI verify
  - Cosign verify for OCI pulls
  - Good first issue: add `--oci-verify` flag to `agent install`

## v0.3.0

- Workflow engine enhancements (branches, retries, conditions)
- Richer telemetry (OpenTelemetry)
- Tool/plugin discovery via registry

