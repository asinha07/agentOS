package main

import (
    "archive/tar"
    "compress/gzip"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/fs"
    "crypto/sha256"
    "bufio"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    cobra "github.com/spf13/cobra"
    "github.com/asinha07/agentOS/pkg/tools"
    "github.com/asinha07/agentOS/pkg/models"
    "github.com/asinha07/agentOS/pkg/workflow"
    regclient "github.com/asinha07/agentOS/pkg/registry/client"
    githubapi "github.com/asinha07/agentOS/pkg/registry/githubapi"
    oci "github.com/asinha07/agentOS/pkg/registry/oci"
)

type Manifest struct {
    Name        string                 `json:"name"`
    Version     string                 `json:"version"`
    Description string                 `json:"description"`
    Entrypoints map[string]any         `json:"entrypoints"`
    Defaults    map[string]any         `json:"defaults"`
    Tools       []any                  `json:"tools"`
    Model       map[string]string      `json:"model"`
    Memory      map[string]any         `json:"memory"`
    Permissions map[string]any         `json:"permissions"`
}

type RunMemory struct {
    dir string
}

func newRunMemory(root string, runID string) (*RunMemory, error) {
    ns := filepath.Join(root, runID)
    if err := os.MkdirAll(ns, 0o755); err != nil {
        return nil, err
    }
    return &RunMemory{dir: ns}, nil
}

func (m *RunMemory) appendEvent(kind string, payload any) error {
    f, err := os.OpenFile(filepath.Join(m.dir, "events.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
    if err != nil {
        return err
    }
    defer f.Close()
    rec := map[string]any{"ts": time.Now().Unix(), "kind": kind, "data": payload}
    b, _ := json.Marshal(rec)
    _, err = f.Write(append(b, '\n'))
    return err
}

func (m *RunMemory) writeKV(key string, value any) error {
    p := filepath.Join(m.dir, "kv.json")
    var data map[string]any
    if b, err := os.ReadFile(p); err == nil {
        if err := json.Unmarshal(b, &data); err != nil {
            data = map[string]any{}
        }
    }
    if data == nil {
        data = map[string]any{}
    }
    data[key] = value
    b, _ := json.MarshalIndent(data, "", "  ")
    return os.WriteFile(p, b, 0o644)
}

// nowID removed (unused)

func loadManifest(dir string) (*Manifest, error) {
    b, err := os.ReadFile(filepath.Join(dir, "agent.yaml"))
    if err != nil {
        return nil, err
    }
    var m Manifest
    if err := json.Unmarshal(b, &m); err != nil {
        return nil, fmt.Errorf("parse agent.yaml as JSON: %w", err)
    }
    if m.Name == "" {
        m.Name = filepath.Base(dir)
    }
    if m.Version == "" {
        m.Version = "0.0.0"
    }
    if err := validateManifest(&m); err != nil { return nil, err }
    return &m, nil
}

func validateManifest(m *Manifest) error {
    if m.Name == "" { return fmt.Errorf("manifest.name required") }
    if m.Version == "" { return fmt.Errorf("manifest.version required") }
    for i, t := range m.Tools {
        switch v := t.(type) {
        case string:
            if v == "" { return fmt.Errorf("tools[%d] empty", i) }
        case map[string]any:
            if name, ok := v["name"].(string); !ok || name == "" { return fmt.Errorf("tools[%d].name required", i) }
        default:
            return fmt.Errorf("tools[%d] has invalid type", i)
        }
    }
    return nil
}

func buildPackage(agentDir string) (string, error) {
    m, err := loadManifest(agentDir)
    if err != nil {
        return "", err
    }
    dist := filepath.Join(agentDir, "dist")
    if err := os.MkdirAll(dist, 0o755); err != nil {
        return "", err
    }
    out := filepath.Join(dist, fmt.Sprintf("%s-%s.agent", m.Name, m.Version))
    f, err := os.Create(out)
    if err != nil {
        return "", err
    }
    defer f.Close()
    gz := gzip.NewWriter(f)
    defer gz.Close()
    tw := tar.NewWriter(gz)
    defer tw.Close()

    err = filepath.WalkDir(agentDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(agentDir, path)
        if rel == "." || strings.HasPrefix(rel, "dist") || strings.HasPrefix(rel, ".git") {
            if d.IsDir() && rel == "dist" {
                return filepath.SkipDir
            }
            return nil
        }
        info, err := d.Info()
        if err != nil {
            return err
        }
        hdr, err := tar.FileInfoHeader(info, "")
        if err != nil {
            return err
        }
        hdr.Name = rel
        if err := tw.WriteHeader(hdr); err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        fi, err := os.Open(path)
        if err != nil {
            return err
        }
        defer fi.Close()
        _, err = io.Copy(tw, fi)
        return err
    })
    if err != nil {
        return "", err
    }
    return out, nil
}

func extractPackage(artifact, dest string) (*Manifest, error) {
    r, err := os.Open(artifact)
    if err != nil {
        return nil, err
    }
    defer r.Close()
    gz, err := gzip.NewReader(r)
    if err != nil {
        return nil, err
    }
    defer gz.Close()
    tr := tar.NewReader(gz)
    if err := os.MkdirAll(dest, 0o755); err != nil {
        return nil, err
    }
    var manifest *Manifest
    for {
        hdr, err := tr.Next()
        if errors.Is(err, io.EOF) {
            break
        }
        if err != nil {
            return nil, err
        }
        target := filepath.Join(dest, hdr.Name)
        if hdr.FileInfo().IsDir() {
            if err := os.MkdirAll(target, 0o755); err != nil {
                return nil, err
            }
            continue
        }
        if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
            return nil, err
        }
        f, err := os.Create(target)
        if err != nil {
            return nil, err
        }
        if _, err := io.Copy(f, tr); err != nil {
            f.Close()
            return nil, err
        }
        f.Close()
        if filepath.Base(target) == "agent.yaml" {
            mf, err := loadManifest(dest)
            if err == nil {
                manifest = mf
            }
        }
    }
    if manifest == nil {
        var err error
        manifest, err = loadManifest(dest)
        if err != nil {
            return nil, fmt.Errorf("no agent.yaml in artifact: %w", err)
        }
    }
    return manifest, nil
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil { return err }
    defer in.Close()
    of, err := os.Create(dst)
    if err != nil { return err }
    if _, err := io.Copy(of, in); err != nil { of.Close(); return err }
    return of.Close()
}

func trySignBlob(path string) error {
    if _, err := exec.LookPath("cosign"); err != nil {
        // Fallback: write SHA256 digest as .sig
        b, _ := os.ReadFile(path)
        sum := sha256Sum(b)
        return os.WriteFile(path+".sig", []byte(sum), 0o644)
    }
    key := os.Getenv("COSIGN_KEY")
    if key == "" { return fmt.Errorf("COSIGN_KEY not set; skipping cosign") }
    sig := path + ".sig"
    cmd := exec.Command("cosign", "sign-blob", "--key", key, path, "--output-signature", sig)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func sha256Sum(b []byte) string {
    h := sha256.New()
    _, _ = h.Write(b)
    return fmt.Sprintf("%x", h.Sum(nil))
}

func runAgent(agentPath string, input string, runsDir string) error {
    return runAgentWithOverrides(agentPath, input, runsDir, "", "")
}

func runAgentWithOverrides(agentPath string, input string, runsDir string, overrideProvider, overrideModel string) error {
    m, err := loadManifest(agentPath)
    if err != nil {
        return err
    }
    if overrideProvider != "" {
        if m.Model == nil { m.Model = map[string]string{} }
        m.Model["provider"] = overrideProvider
    }
    if overrideModel != "" {
        if m.Model == nil { m.Model = map[string]string{} }
        m.Model["model"] = overrideModel
    }
    // Try to load workflow.yaml (JSON content).
    var wf *workflow.WorkflowSpec
    if b, err := os.ReadFile(filepath.Join(agentPath, "workflow.yaml")); err == nil {
        var spec workflow.WorkflowSpec
        if json.Unmarshal(b, &spec) == nil && len(spec.Steps) > 0 {
            wf = &spec
        }
    }
    runID := strings.ReplaceAll(time.Now().UTC().Format("20060102-150405.000000Z07:00"), ":", "")
    mem, err := newRunMemory(runsDir, runID)
    if err != nil {
        return err
    }
    topic := input
    if topic == "" {
        if v, ok := m.Defaults["input"].(string); ok {
            topic = v
        }
    }
    _ = mem.appendEvent("start", map[string]any{"agent": m.Name, "version": m.Version, "topic": topic, "provider": m.Model["provider"], "model": m.Model["model"]})

    // Special composite agent: app-team
    if m.Name == "app-team" {
        agentsDir := filepath.Dir(agentPath) // points to agents/
        team := []string{"product-manager", "be-developer", "web-developer", "qa", "code-reviewer"}
        fmt.Printf("AgentOS — app-team\n        Run ID: %s\n\n", runID)
        for _, member := range team {
            path := filepath.Join(agentsDir, member)
            fmt.Printf("→ Running %s\n", member)
            if err := runAgentWithOverrides(path, topic, runsDir, overrideProvider, overrideModel); err != nil {
                return fmt.Errorf("member %s failed: %w", member, err)
            }
        }
        _ = mem.appendEvent("final", map[string]any{"output": "team done"})
        return nil
    }
    // Built-in tools
    // ideas placeholder for future ideation step
    best := topic
    // Provider selection
    provider := strings.ToLower(m.Model["provider"])
    modelID := m.Model["model"]

    // Execute optional declared tools (web_search, file_reader, http_client)
    perms := m.Permissions
    ctx := tools.Context{Internet: toBool(perms["internet"]), Filesystem: toFS(perms["filesystem"]), Workdir: mustGetwd()}
    // web_search (optional)
    researchQuery := topic + " competitors"
    if wf != nil {
        for _, s := range wf.Steps {
            if s.Type == "research" && s.Query != "" {
                q := s.Query
                q = strings.ReplaceAll(q, "{idea}", topic)
                researchQuery = q
                break
            }
        }
    }
    if hasTool(m.Tools, "web_search") {
        if t, ok := tools.Get("web_search"); ok {
            out, _ := t.Execute(map[string]any{"query": researchQuery}, ctx)
            _ = mem.appendEvent("tool.web_search", out)
        }
    }
    if hasTool(m.Tools, "file_reader") {
        // Read defaults.path if present
        path, _ := m.Defaults["path"].(string)
        if path != "" {
            if t, ok := tools.Get("file_reader"); ok {
                if out, err := t.Execute(map[string]any{"path": path}, ctx); err == nil {
                    _ = mem.appendEvent("tool.file_reader", out)
                }
            }
        }
    }
    // http_client: only if url in defaults and internet allowed
    var httpStatus int
    if hasTool(m.Tools, "http_client") {
        url, _ := m.Defaults["url"].(string)
        if url != "" {
            if t, ok := tools.Get("http_client"); ok {
                if out, err := t.Execute(map[string]any{"url": url}, ctx); err == nil {
                    httpStatus, _ = out["status"].(int)
                    _ = mem.appendEvent("tool.http_client", out)
                } else {
                    _ = mem.appendEvent("tool.http_client.error", map[string]any{"error": err.Error()})
                }
            }
        }
    }

    // If workflow present, execute its steps to shape output accordingly.
    fmt.Printf("AgentOS — %s\n", m.Name)
    fmt.Printf("        Run ID: %s\n", runID)
    activeProvider := "mock"
    activeModel := "mock"
    fmt.Printf("Context: %s\n\n", topic)

    // Helper to generate content via selected provider
    genWithProvider := func(prompt string) (string, bool) {
        switch provider {
        case "openai":
            if os.Getenv("OPENAI_API_KEY") != "" {
                if content, err := (models.OpenAI{Model: modelID}).Generate(prompt); err == nil {
                    activeProvider, activeModel = "openai", modelID
                    return content, true
                }
            }
        case "anthropic", "claude":
            if os.Getenv("ANTHROPIC_API_KEY") != "" {
                if content, err := (models.Anthropic{Model: modelID}).Generate(prompt); err == nil {
                    activeProvider, activeModel = "anthropic", modelID
                    return content, true
                }
            }
        case "xai", "grok":
            if os.Getenv("XAI_API_KEY") != "" {
                if content, err := (models.Grok{Model: modelID}).Generate(prompt); err == nil {
                    activeProvider, activeModel = "xai", modelID
                    return content, true
                }
            }
        }
        return "", false
    }

    // Team agents: generate role-specific documents
    outPath := "artifact.md"
    if v, ok := m.Defaults["output"].(string); ok && v != "" { outPath = v }
    var wrote string
    if hasTool(m.Tools, "file_writer") && (m.Name == "product-manager" || m.Name == "be-developer" || m.Name == "web-developer" || m.Name == "qa" || m.Name == "code-reviewer") {
        // Prompt base from prompt.md if present
        basePrompt := ""
        if b, err := os.ReadFile(filepath.Join(agentPath, "prompt.md")); err == nil { basePrompt = string(b) }
        var content string
        switch m.Name {
        case "product-manager":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nPlease write a concise PRD covering problem, personas, features + acceptance criteria, success metrics, risks/assumptions, milestones.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# PRD — %s\n\n## Problem\n...\n\n## Personas\n...\n\n## Features\n1. ...\n2. ...\n\n## Metrics\n...\n\n## Risks\n...\n\n## Milestones\n...\n", topic)
            }
        case "be-developer":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a backend design: APIs (paths, payloads), data model, services, NFRs, security.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Backend Design — %s\n\n## APIs\n- GET /...\n\n## Data Model\n- tables and relationships\n\n## Services\n- responsibilities\n\n## NFRs\n- perf, reliability, observability\n\n## Security\n- auth, encryption, secrets\n", topic)
            }
        case "web-developer":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a frontend design: architecture, routes/components, flows, quality (a11y/perf/i18n).", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Frontend Design — %s\n\n## Architecture\n- routing, state management\n\n## Routes\n- / -> ...\n\n## UI Flows\n- ...\n\n## Quality\n- a11y, perf, i18n\n", topic)
            }
        case "qa":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a QA test plan: scope, scenarios (happy/edge), integration/performance tests, data, environments.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Test Plan — %s\n\n## Scope\n...\n\n## Scenarios\n- happy paths\n- edge cases\n\n## Integration & Performance\n...\n", topic)
            }
        case "code-reviewer":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a review: code-review checklist, design notes (BE/FE), risks and follow-ups.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Review — %s\n\n## Checklist\n- correctness, clarity, safety\n\n## Design Notes\n- backend and frontend\n\n## Risks & Follow-ups\n...\n", topic)
            }
        }
        if t, ok := tools.Get("file_writer"); ok {
            out, err := t.Execute(map[string]any{"path": outPath, "content": content}, ctx)
            if err == nil {
                _ = mem.appendEvent("tool.file_writer", out)
                if p, ok := out["path"].(string); ok { wrote = p }
            } else {
                _ = mem.appendEvent("tool.file_writer.error", map[string]any{"error": err.Error()})
            }
        }
    }
    // Generic content (and model) for other agent names when file_writer is available
    if wrote == "" && hasTool(m.Tools, "file_writer") {
        // Read prompt.md if present
        basePrompt := ""
        if b, err := os.ReadFile(filepath.Join(agentPath, "prompt.md")); err == nil { basePrompt = string(b) }
        var content string
        switch m.Name {
        case "coding-agent":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a coding plan: tasks, guidelines, tests.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Coding Plan — %s\n\n## Tasks\n- scaffold project\n- implement modules\n- write tests\n\n## Guidelines\n- code style\n- error handling\n- logging\n", topic)
            }
        case "research-agent":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a research report: summary and sources.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Research Report — %s\n\n## Summary\n...\n\n## Sources\n- ...\n", topic)
            }
        case "seo-agent":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite an SEO brief: keywords, content plan, technical checklist, measurement.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# SEO Brief — %s\n\n## Keywords\n- ...\n\n## Content Plan\n- ...\n\n## Technical\n- sitemap, robots, meta\n", topic)
            }
        case "growth-hacker":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a growth plan: channels, experiments, loops, metrics.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Growth Plan — %s\n\n## Channels\n- social, content, partnerships\n\n## Experiments\n- A/B ideas\n\n## Metrics\n- activation, retention, referral\n", topic)
            }
        case "bug-fixer":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a bug fix plan: reproduction, root cause, fix & tests, rollout.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Bug Fix Plan — %s\n\n## Reproduction\n...\n\n## Root Cause Analysis\n...\n\n## Fix & Tests\n...\n", topic)
            }
        case "data-analyst":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite an analysis plan: questions, sources, methods, viz.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Analysis Plan — %s\n\n## Questions\n- ...\n\n## Data Sources\n- ...\n\n## Methods\n- SQL, stats, viz\n", topic)
            }
        case "web-scraper":
            prompt := fmt.Sprintf("%s\n\nTopic: %s\nWrite a scraping outline: targets, selectors, storage, anti-bot.", basePrompt, topic)
            if c, ok := genWithProvider(prompt); ok { content = c } else {
                content = fmt.Sprintf("# Scraping Outline — %s\n\n## Targets\n- URLs\n\n## Selectors\n- CSS/XPath\n\n## Storage\n- CSV/DB\n", topic)
            }
        }
        if content != "" {
            if v, ok := m.Defaults["output"].(string); ok && v != "" { outPath = v } else { outPath = "artifact.md" }
            if t, ok := tools.Get("file_writer"); ok {
                out, err := t.Execute(map[string]any{"path": outPath, "content": content}, ctx)
                if err == nil {
                    _ = mem.appendEvent("tool.file_writer", out)
                    if p, ok := out["path"].(string); ok { wrote = p }
                } else {
                    _ = mem.appendEvent("tool.file_writer.error", map[string]any{"error": err.Error()})
                }
            }
        }
    }
    if wrote != "" { fmt.Printf("Wrote %s\n", wrote) }
    if httpStatus != 0 { fmt.Printf("HTTP status: %d\n", httpStatus) }
    // Emit model info event with actual provider used
    _ = mem.appendEvent("model.info", map[string]any{"active_provider": activeProvider, "active_model": activeModel})
    _ = mem.appendEvent("final", map[string]any{"output": "ok"})
    _ = mem.writeKV("topic", topic)
    _ = mem.writeKV("best_idea", best)
    return nil
}

func toBool(v any) bool {
    b, _ := v.(bool)
    return b
}
func toFS(v any) bool {
    switch vv := v.(type) {
    case bool:
        return vv
    case string:
        return vv == "limited" || vv == "true" || vv == "rw"
    default:
        return false
    }
}
func hasTool(ts []any, name string) bool {
    for _, t := range ts {
        switch v := t.(type) {
        case string:
            if v == name { return true }
        case map[string]any:
            if vv, ok := v["name"].(string); ok && vv == name { return true }
        }
    }
    return false
}
func mustGetwd() string { wd, _ := os.Getwd(); return wd }

func main() {
    var runsDir = "runs"
    var builtins = "agents"
    var examples = "examples"
    var installed = "installed_agents"

    root := &cobra.Command{Use: "agent", Short: "AgentOS CLI"}

    // run
    var input string
    var overrideProvider string
    var overrideModel string
    var runRegistry string
    run := &cobra.Command{Use: "run", Short: "Run an agent", RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("agent name or path required")
        }
        ref := args[0]
        if runRegistry == "" { runRegistry = os.Getenv("AGENT_REGISTRY") }
        // Resolve: installed -> built-in (agents/examples) -> directory -> artifact
        var dir string
        candidates := []string{filepath.Join(installed, ref), filepath.Join(builtins, ref), filepath.Join(examples, ref), ref}
        for _, c := range candidates {
            if fi, err := os.Stat(c); err == nil && fi.IsDir() {
                dir = c
                break
            }
        }
        if dir == "" && strings.HasSuffix(ref, ".agent") {
            // install to installed/<name>
            tmp := filepath.Join(installed, strings.TrimSuffix(filepath.Base(ref), ".agent"))
            if _, err := extractPackage(ref, tmp); err != nil {
                return err
            }
            dir = tmp
        }
        // Auto-install from registry if not found
        if dir == "" && runRegistry != "" && !strings.HasSuffix(ref, ".agent") {
            results, err := regclient.Search(runRegistry, ref)
            if err != nil { return err }
            if len(results) > 0 {
                file := results[0].File
                dl, err := regclient.Download(runRegistry, file)
                if err != nil { return err }
                dest := filepath.Join(installed, results[0].Name)
                os.RemoveAll(dest)
                if _, err := extractPackage(dl, dest); err != nil { return err }
                dir = dest
                fmt.Printf("Pulled %s@%s from %s\n", results[0].Name, results[0].Version, runRegistry)
            }
        }
        if dir == "" {
            return fmt.Errorf("agent '%s' not found", ref)
        }
        // Prompt for input if not provided
        if input == "" {
            fmt.Print("What startup idea do you want to explore? ")
            reader := bufio.NewReader(os.Stdin)
            line, _ := reader.ReadString('\n')
            input = strings.TrimSpace(line)
        }
        return runAgentWithOverrides(dir, input, runsDir, overrideProvider, overrideModel)
    }}
    run.Flags().StringVar(&input, "input", "", "Optional input for the agent")
    run.Flags().StringVar(&runRegistry, "registry", "", "Registry base URL to auto-install missing agents")
    run.Flags().StringVar(&overrideProvider, "provider", "", "Override model provider (openai|anthropic|xai)")
    run.Flags().StringVar(&overrideModel, "model", "", "Override model id (e.g., gpt-4.1, claude-3-5-sonnet-latest, grok-2)")

    // build
    var buildPath string
    build := &cobra.Command{Use: "build", Short: "Build a .agent package", RunE: func(cmd *cobra.Command, args []string) error {
        p := buildPath
        if p == "" {
            if len(args) > 0 {
                p = args[0]
            } else {
                p = "."
            }
        }
        out, err := buildPackage(p)
        if err != nil {
            return err
        }
        fmt.Println(out)
        return nil
    }}
    build.Flags().StringVar(&buildPath, "path", "", "Path to agent directory")

    // install
    var regURL string
    var artifactURL string
    var ociPullRef string
    var ghRef string
    install := &cobra.Command{Use: "install", Short: "Install an agent", RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            // list built-ins
            entries, _ := os.ReadDir(builtins)
            fmt.Println("Built-in agents:")
            for _, e := range entries {
                if _, err := os.Stat(filepath.Join(builtins, e.Name(), "agent.yaml")); err == nil {
                    fmt.Println("-", e.Name())
                }
            }
            return nil
        }
        src := args[0]
        // From registry by name
        if ghRef != "" || strings.HasPrefix(src, "github.com/") || (!strings.Contains(src, "/") && strings.Count(src, "/")==0 && strings.Contains(src, "@") && strings.Count(src, "@")==1) {
            // GitHub ref: github.com/owner/repo@tag or owner/repo@tag
            ref := src
            if ghRef != "" { ref = ghRef }
            dl, err := githubapi.InstallRef(ref)
            if err != nil { return err }
            mf, err := extractPackage(dl, installed)
            if err != nil { return err }
            fmt.Printf("Installed %s@%s from GitHub\n", mf.Name, mf.Version)
            return nil
        }
        if regURL != "" && !strings.HasSuffix(src, ".agent") && !strings.Contains(src, "/") {
            results, err := regclient.Search(regURL, src)
            if err != nil { return err }
            if len(results) == 0 { return fmt.Errorf("no results for %s", src) }
            file := results[0].File
            dl, err := regclient.Download(regURL, file)
            if err != nil { return err }
            mf, err := extractPackage(dl, installed)
            if err != nil { return err }
            fmt.Printf("Installed %s@%s\n", mf.Name, mf.Version)
            return nil
        }
        // From direct artifact URL
        if artifactURL != "" {
            dl, err := regclient.Download(artifactURL, "")
            if err != nil { return err }
            mf, err := extractPackage(dl, installed)
            if err != nil { return err }
            fmt.Printf("Installed %s@%s\n", mf.Name, mf.Version)
            return nil
        }
        if ociPullRef != "" {
            dl, err := oci.Pull(ociPullRef)
            if err != nil { return err }
            mf, err := extractPackage(dl, installed)
            if err != nil { return err }
            fmt.Printf("Installed %s@%s from OCI\n", mf.Name, mf.Version)
            return nil
        }
        if strings.HasSuffix(src, ".agent") {
            mf, err := extractPackage(src, installed)
            if err != nil {
                return err
            }
            fmt.Printf("Installed %s@%s\n", mf.Name, mf.Version)
            return nil
        }
        // allow built-in names
        if _, err := os.Stat(filepath.Join(builtins, src)); err == nil {
            src = filepath.Join(builtins, src)
        } else if _, err := os.Stat(filepath.Join(examples, src)); err == nil {
            src = filepath.Join(examples, src)
        }
        // copy directory
        name := filepath.Base(src)
        dest := filepath.Join(installed, name)
        os.RemoveAll(dest)
        return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }
            rel, _ := filepath.Rel(src, path)
            if rel == "." {
                return nil
            }
            target := filepath.Join(dest, rel)
            if d.IsDir() {
                return os.MkdirAll(target, 0o755)
            }
            if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
                return err
            }
            in, err := os.Open(path)
            if err != nil {
                return err
            }
            defer in.Close()
            out, err := os.Create(target)
            if err != nil {
                return err
            }
            if _, err := io.Copy(out, in); err != nil {
                out.Close()
                return err
            }
            return out.Close()
        })
    }}
    install.Flags().StringVar(&regURL, "registry", "", "Registry base URL (e.g. http://localhost:8080)")
    install.Flags().StringVar(&artifactURL, "url", "", "Direct artifact URL (overrides name)")
    install.Flags().StringVar(&ociPullRef, "oci-ref", "", "OCI reference to pull (requires oras CLI)")
    install.Flags().StringVar(&ghRef, "github", "", "GitHub ref owner/repo[@tag] (overrides name)")

    // inspect
    inspect := &cobra.Command{Use: "inspect", Short: "Inspect an agent", RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("agent name or artifact required")
        }
        ref := args[0]
        if strings.HasSuffix(ref, ".agent") {
            tmp := filepath.Join(".tmp_inspect")
            os.RemoveAll(tmp)
            mf, err := extractPackage(ref, tmp)
            if err != nil {
                return err
            }
            b, _ := json.MarshalIndent(mf, "", "  ")
            fmt.Println(string(b))
            os.RemoveAll(tmp)
            return nil
        }
        dir := filepath.Join(installed, ref)
        if _, err := os.Stat(filepath.Join(dir, "agent.yaml")); err != nil {
            dir = filepath.Join(builtins, ref)
            if _, err := os.Stat(filepath.Join(dir, "agent.yaml")); err != nil {
                dir = filepath.Join(examples, ref)
            }
        }
        mf, err := loadManifest(dir)
        if err != nil {
            return err
        }
        b, _ := json.MarshalIndent(mf, "", "  ")
        fmt.Println(string(b))
        return nil
    }}

    // logs
    var tail bool
    var limit int
    logs := &cobra.Command{Use: "logs", Short: "Show recent runs", RunE: func(cmd *cobra.Command, args []string) error {
        entries, _ := os.ReadDir(runsDir)
        var names []string
        for _, e := range entries {
            if e.IsDir() {
                names = append(names, e.Name())
            }
        }
        if len(names) == 0 {
            fmt.Println("No runs found")
            return nil
        }
        // naive: show latest dir
        if tail {
            latest := names[len(names)-1]
            b, _ := os.ReadFile(filepath.Join(runsDir, latest, "events.jsonl"))
            fmt.Print(string(b))
            return nil
        }
        for i, n := range names {
            fmt.Println(n)
            if i+1 >= limit && limit > 0 {
                break
            }
        }
        return nil
    }}
    logs.Flags().BoolVar(&tail, "tail", false, "Tail latest run")
    logs.Flags().IntVar(&limit, "limit", 5, "Limit number of runs displayed")

    // init (skeleton)
    initCmd := &cobra.Command{Use: "init", Short: "Create a new agent skeleton", RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("agent name required")
        }
        name := args[0]
        base := filepath.Join("installed_agents", name)
        if err := os.MkdirAll(filepath.Join(base, "tools"), 0o755); err != nil {
            return err
        }
        mf := Manifest{Name: name, Version: "0.1.0", Description: "Example agent", Defaults: map[string]any{"input": "Hello"}, Tools: []any{}, Model: map[string]string{"provider": "mock", "model": "mock-001"}, Memory: map[string]any{"type": "jsonl"}}
        b, _ := json.MarshalIndent(mf, "", "  ")
        if err := os.WriteFile(filepath.Join(base, "agent.yaml"), b, 0o644); err != nil {
            return err
        }
        return os.WriteFile(filepath.Join(base, "prompt.md"), []byte("You are a helpful agent."), 0o644)
    }}

    // publish (local FS or OCI)
    var ociRef string
    var signBlob bool
    publish := &cobra.Command{Use: "publish", Short: "Publish agent (local folder or OCI)", RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("agent name required")
        }
        dir := filepath.Join(installed, args[0])
        if _, err := os.Stat(filepath.Join(dir, "agent.yaml")); err != nil {
            // Try built-ins or examples
            d1 := filepath.Join(builtins, args[0])
            d2 := filepath.Join(examples, args[0])
            if _, e1 := os.Stat(filepath.Join(d1, "agent.yaml")); e1 == nil {
                dir = d1
            } else if _, e2 := os.Stat(filepath.Join(d2, "agent.yaml")); e2 == nil {
                dir = d2
            } else {
                return fmt.Errorf("agent not found: %s", args[0])
            }
        }
        out, err := buildPackage(dir)
        if err != nil {
            return err
        }
        if signBlob {
            // Optional: sign blob via cosign if available; write .sig next to artifact
            if err := trySignBlob(out); err != nil {
                fmt.Fprintln(os.Stderr, "Warning: signing failed:", err)
            }
        }
        if ociRef != "" {
            // Push to OCI via oras CLI
            if err := oci.Push(ociRef, out); err != nil { return err }
            fmt.Println("Pushed to OCI:", ociRef)
            return nil
        }
        // Default: local registry folder
        reg := filepath.Join("registry", "agents")
        if err := os.MkdirAll(reg, 0o755); err != nil { return err }
        dst := filepath.Join(reg, filepath.Base(out))
        if err := copyFile(out, dst); err != nil { return err }
        fmt.Println("Published:", dst)
        return nil
    }}
    publish.Flags().StringVar(&ociRef, "oci-ref", "", "OCI reference to push (requires oras CLI)")
    publish.Flags().BoolVar(&signBlob, "sign", false, "Sign artifact with cosign if available")

    // compose (simple sequential runner)
    var composeAgents string
    var composeInput string
    compose := &cobra.Command{Use: "compose", Short: "Compose multiple agents sequentially", RunE: func(cmd *cobra.Command, args []string) error {
        if composeAgents == "" { return fmt.Errorf("--agents required (comma-separated)") }
        names := strings.Split(composeAgents, ",")
        for _, n := range names {
            n = strings.TrimSpace(n)
            if n == "" { continue }
            if err := runAgent(filepath.Join(builtins, n), composeInput, runsDir); err != nil {
                return err
            }
        }
        return nil
    }}
    compose.Flags().StringVar(&composeAgents, "agents", "", "Comma-separated agent names")
    compose.Flags().StringVar(&composeInput, "input", "", "Input passed to each agent")

    // search
    var searchRegistry string
    var searchGitHub bool
    search := &cobra.Command{Use: "search", Short: "Search registry or GitHub for agents", RunE: func(cmd *cobra.Command, args []string) error {
        if !searchGitHub && searchRegistry == "" {
            if os.Getenv("AGENT_REGISTRY") != "" {
                searchRegistry = os.Getenv("AGENT_REGISTRY")
            } else {
                return fmt.Errorf("set --registry or use --github to search GitHub")
            }
        }
        q := ""
        if len(args) > 0 { q = args[0] }
        if searchGitHub {
            repos, err := githubapi.SearchRepos(q)
            if err != nil { return err }
            for _, r := range repos { fmt.Printf("%s — %s\n", r.FullName, r.Description) }
            return nil
        } else {
            results, err := regclient.Search(searchRegistry, q)
            if err != nil { return err }
            for _, a := range results {
                fmt.Printf("%s@%s (%s)\n", a.Name, a.Version, a.File)
            }
            return nil
        }
    }}
    search.Flags().StringVar(&searchRegistry, "registry", "", "Registry base URL (defaults to $AGENT_REGISTRY)")
    search.Flags().BoolVar(&searchGitHub, "github", false, "Search GitHub (topic: agentos-agent)")

    // team alias for app-team
    team := &cobra.Command{Use: "team", Short: "Run the five-agent team", RunE: func(cmd *cobra.Command, args []string) error {
        return runAgentWithOverrides(filepath.Join(builtins, "app-team"), input, runsDir, overrideProvider, overrideModel)
    }}
    team.Flags().StringVar(&input, "input", "", "Optional input for the team")
    team.Flags().StringVar(&overrideProvider, "provider", "", "Override model provider (openai|anthropic|xai)")
    team.Flags().StringVar(&overrideModel, "model", "", "Override model id")

    // doctor
    var doctorRegistry string
    doctor := &cobra.Command{Use: "doctor", Short: "Check environment and provider connectivity", RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("AgentOS Doctor — environment checks")
        // Providers
        // OpenAI
        if os.Getenv("OPENAI_API_KEY") == "" {
            fmt.Println("OpenAI: missing OPENAI_API_KEY (optional)")
        } else {
            fmt.Print("OpenAI: testing ... ")
            if _, err := (models.OpenAI{Model: "gpt-4.1"}).Generate("ping"); err != nil {
                fmt.Printf("ERROR (%v)\n", err)
            } else {
                fmt.Println("OK")
            }
        }
        // Anthropic
        if os.Getenv("ANTHROPIC_API_KEY") == "" {
            fmt.Println("Anthropic: missing ANTHROPIC_API_KEY (optional)")
        } else {
            fmt.Print("Anthropic: testing ... ")
            if _, err := (models.Anthropic{Model: "claude-3-5-sonnet-latest"}).Generate("ping"); err != nil {
                fmt.Printf("ERROR (%v)\n", err)
            } else {
                fmt.Println("OK")
            }
        }
        // xAI
        if os.Getenv("XAI_API_KEY") == "" {
            fmt.Println("xAI (Grok): missing XAI_API_KEY (optional)")
        } else {
            fmt.Print("xAI (Grok): testing ... ")
            if _, err := (models.Grok{Model: "grok-2"}).Generate("ping"); err != nil {
                fmt.Printf("ERROR (%v)\n", err)
            } else {
                fmt.Println("OK")
            }
        }
        // GitHub token (rate limit)
        if os.Getenv("GITHUB_TOKEN") == "" {
            fmt.Println("GitHub: missing GITHUB_TOKEN (optional; improves rate limits)")
        } else {
            fmt.Print("GitHub API: testing rate_limit ... ")
            req, _ := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
            req.Header.Set("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))
            req.Header.Set("Accept", "application/vnd.github+json")
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                fmt.Printf("ERROR (%v)\n", err)
            } else {
                defer resp.Body.Close()
                if resp.StatusCode == 200 {
                    fmt.Println("OK")
                } else {
                    fmt.Printf("status %d\n", resp.StatusCode)
                }
            }
        }
        // Registry
        reg := doctorRegistry
        if reg == "" { reg = os.Getenv("AGENT_REGISTRY") }
        if reg != "" {
            fmt.Printf("Registry: testing %s ... ", reg)
            url := reg
            if !strings.Contains(url, "/search") {
                if strings.HasSuffix(url, "/") { url = url+"search" } else { url = url+"/search" }
            }
            req, _ := http.NewRequest("GET", url+"?q=doctor", nil)
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                fmt.Printf("ERROR (%v)\n", err)
            } else {
                defer resp.Body.Close()
                if resp.StatusCode == 200 { fmt.Println("OK") } else { fmt.Printf("status %d\n", resp.StatusCode) }
            }
        } else {
            fmt.Println("Registry: not set (set AGENT_REGISTRY or use --registry)")
        }
        fmt.Println("Doctor checks completed.")
        return nil
    }}
    doctor.Flags().StringVar(&doctorRegistry, "registry", "", "Registry URL to test (defaults to $AGENT_REGISTRY)")

    root.AddCommand(run, build, install, inspect, logs, initCmd, publish, compose, search, team, doctor)
    if err := root.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(2)
    }
}
