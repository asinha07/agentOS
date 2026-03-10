package tools

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    "golang.org/x/text/cases"
    "golang.org/x/text/language"
)

type Context struct {
    Internet bool
    Filesystem bool
    Workdir string
}

type Result map[string]any

type Tool interface {
    Name() string
    Execute(input map[string]any, ctx Context) (Result, error)
    Schema() map[string]any
    Metadata() map[string]any
}

// Registry
var registry = map[string]Tool{}

func Register(t Tool) { registry[t.Name()] = t }

func Get(name string) (Tool, bool) { t, ok := registry[name]; return t, ok }

// Built-in: web_search (stub)
type WebSearch struct{}

func (w WebSearch) Name() string { return "web_search" }
func (w WebSearch) Execute(input map[string]any, ctx Context) (Result, error) {
    q, _ := input["query"].(string)
    if q == "" { q = "example" }
    // Stub results only; no network used
    title := cases.Title(language.Und)
    results := []map[string]string{
        {"title": fmt.Sprintf("%s — Overview", title.String(q)), "url": fmt.Sprintf("https://example.com/%s", strings.ReplaceAll(q, " ", "-"))},
        {"title": fmt.Sprintf("%s — Guides", title.String(q)), "url": fmt.Sprintf("https://example.com/%s-guides", strings.ReplaceAll(q, " ", "-"))},
    }
    return Result{"results": results}, nil
}
func (w WebSearch) Schema() map[string]any { return map[string]any{"input_schema": map[string]any{"query": "string"}} }
func (w WebSearch) Metadata() map[string]any { return map[string]any{"transport": "builtin"} }

// Built-in: file_reader
type FileReader struct{}
func (f FileReader) Name() string { return "file_reader" }
func (f FileReader) Execute(input map[string]any, ctx Context) (Result, error) {
    if !ctx.Filesystem { return nil, errors.New("filesystem access denied") }
    p, _ := input["path"].(string)
    if p == "" { return nil, errors.New("path required") }
    if !strings.HasPrefix(p, "/") {
        p = ctx.Workdir + string(os.PathSeparator) + p
    }
    // Restrict to workdir
    if !strings.HasPrefix(p, ctx.Workdir) {
        return nil, errors.New("path outside workdir")
    }
    b, err := os.ReadFile(p)
    if err != nil { return nil, err }
    // Limit size
    if len(b) > 64*1024 { b = b[:64*1024] }
    return Result{"path": p, "content": string(b)}, nil
}
func (f FileReader) Schema() map[string]any { return map[string]any{"input_schema": map[string]any{"path": "string"}} }
func (f FileReader) Metadata() map[string]any { return map[string]any{"transport": "builtin"} }

// Built-in: http_client (GET)
type HttpClient struct{}
func (h HttpClient) Name() string { return "http_client" }
func (h HttpClient) Execute(input map[string]any, ctx Context) (Result, error) {
    if !ctx.Internet { return nil, errors.New("internet access denied") }
    url, _ := input["url"].(string)
    if url == "" { return nil, errors.New("url required") }
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(url)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 { return nil, fmt.Errorf("http status %d", resp.StatusCode) }
    b, _ := io.ReadAll(resp.Body)
    if len(b) > 64*1024 { b = b[:64*1024] }
    return Result{"status": resp.StatusCode, "body": string(b)}, nil
}
func (h HttpClient) Schema() map[string]any { return map[string]any{"input_schema": map[string]any{"url": "string"}} }
func (h HttpClient) Metadata() map[string]any { return map[string]any{"transport": "http"} }

func init() {
    Register(WebSearch{})
    Register(FileReader{})
    Register(HttpClient{})
}

// Built-in: file_writer
type FileWriter struct{}

func (w FileWriter) Name() string { return "file_writer" }
func (w FileWriter) Execute(input map[string]any, ctx Context) (Result, error) {
    if !ctx.Filesystem { return nil, errors.New("filesystem access denied") }
    p, _ := input["path"].(string)
    content, _ := input["content"].(string)
    if p == "" { return nil, errors.New("path required") }
    if !strings.HasPrefix(p, "/") {
        p = ctx.Workdir + string(os.PathSeparator) + p
    }
    if !strings.HasPrefix(p, ctx.Workdir) { return nil, errors.New("path outside workdir") }
    if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil { return nil, err }
    if err := os.WriteFile(p, []byte(content), 0o644); err != nil { return nil, err }
    return Result{"path": p, "bytes": len(content)}, nil
}
func (w FileWriter) Schema() map[string]any { return map[string]any{"input_schema": map[string]any{"path": "string", "content": "string"}} }
func (w FileWriter) Metadata() map[string]any { return map[string]any{"transport": "builtin"} }

func init() {
    Register(FileWriter{})
}
