Tools & Permissions

Built-ins
- web_search: returns competitor-like results (offline stub)
  - perms: none (mocked)
- http_client: HTTP GET for a URL
  - perms: `internet: true`
- file_reader: reads a file within the working directory
  - perms: `filesystem: limited`
- file_writer: writes content to a file within the working directory
  - perms: `filesystem: limited`

Usage
- Declare tool names in `agent.yaml` under `tools: []`.
- The runtime enforces permissions before executing a tool.

Roadmap
- Per-tool scopes, rate limits, and concurrency controls.
- Discovery via registries and third-party tool catalogs.

