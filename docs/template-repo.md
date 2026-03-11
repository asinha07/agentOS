Template: Agent Repository

Goal
- Make it trivial to publish an agent to GitHub Releases as a `.agent` artifact so anyone can install with:
  - `agent install github.com/<owner>/<repo>@<tag>`

Repo Layout
```
my-agent/
  agent.yaml
  prompt.md
  README.md
  .github/workflows/release.yml
```

1) Minimal agent.yaml
```
{
  "name": "my-agent",
  "version": "1.0.0",
  "description": "What your agent does.",
  "entrypoints": {"run": {"type": "builtin"}},
  "defaults": {"input": "Hello world", "output": "artifact.md"},
  "tools": ["file_writer"],
  "model": {"provider": "openai", "model": "gpt-4.1"},
  "memory": {"type": "jsonl"},
  "permissions": {"filesystem": "limited"}
}
```

2) Minimal prompt.md
```
You are a helpful agent. Write a short artifact for the given input.
```

3) GitHub Actions (release.yml)
Use macOS runner to install AgentOS via Homebrew and build `.agent`:
```
name: Release
on:
  push:
    tags: [ 'v*' ]

jobs:
  release:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      - run: brew tap asinha07/homebrew-tap && brew install agent
      - run: agent build .
      - uses: softprops/action-gh-release@v2
        with:
          files: |
            dist/*.agent
```

4) Publish
- Tag a release: `git tag v1.0.0 && git push origin v1.0.0`
- Your `.agent` file appears in Releases.

5) Install
- Consumers install directly from GitHub:
  - `agent install github.com/<owner>/<repo>@v1.0.0`

Tips
- Add the repo topic `agentos-agent` so `agent search --github` can find it.
- Include a short README with usage and an example output.

