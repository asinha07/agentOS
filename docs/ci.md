CI & Release Setup

Secrets
- GH_PAT (required): Classic PAT with `repo` scope for Homebrew tap pushes (GoReleaser uses it via `BREW_TAP_TOKEN`).
- GH_ADMIN_PAT (optional): Classic PAT with `repo` scope to set the repository topic via the `set-topic` workflow.

Permissions
- Actions → General → Workflow permissions: Read and write

Workflows
- ci.yml: builds on Go 1.21/1.22, vet + lint, and warns if `go mod tidy` would change files.
- release.yml: runs GoReleaser, then builds and uploads `.agent` artifacts to the GitHub Release; also deletes an existing release for the tag to avoid asset conflicts.
- pages.yml: publishes `docs/` to GitHub Pages.
- set-topic.yml: adds the `agentos-agent` topic on release (needs GH_ADMIN_PAT).

Releasing
- Tag: `git tag vX.Y.Z && git push origin vX.Y.Z`
- GoReleaser publishes binaries, deb/rpm, SBOM, SLSA, and opens a Homebrew PR.
- The workflow then packages built-in agents and attaches them to the release.

Notes
- For the cleanest releases, commit `go.mod`/`go.sum` after running `go mod tidy` locally.

