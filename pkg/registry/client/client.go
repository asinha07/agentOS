package client

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "time"
    "crypto/sha256"
)

type AgentInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    File    string `json:"file"`
}

func Search(base, q string) ([]AgentInfo, error) {
    u, _ := url.Parse(base)
    u.Path = strings.TrimSuffix(u.Path, "/") + "/search"
    qs := u.Query()
    qs.Set("q", q)
    u.RawQuery = qs.Encode()
    req, _ := http.NewRequest("GET", u.String(), nil)
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return nil, fmt.Errorf("registry search status %d", resp.StatusCode) }
    var out struct{ Agents []AgentInfo `json:"agents"` }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
    return out.Agents, nil
}

func Download(base, file string) (string, error) {
    u := strings.TrimSuffix(base, "/") + "/agents/" + file
    req, _ := http.NewRequest("GET", u, nil)
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return "", fmt.Errorf("registry download status %d", resp.StatusCode) }
    os.MkdirAll(".downloads", 0o755)
    out := filepath.Join(".downloads", file)
    f, err := os.Create(out)
    if err != nil { return "", err }
    defer f.Close()
    if _, err := io.Copy(f, resp.Body); err != nil { return "", err }
    // Try to verify against .sig (SHA256 hex) if present
    sigURL := u + ".sig"
    if sig, err := fetch(sigURL); err == nil && len(sig) > 0 {
        if err := verifySHA256(out, strings.TrimSpace(string(sig))); err != nil {
            return "", fmt.Errorf("signature mismatch for %s: %w", out, err)
        }
    }
    return out, nil
}

// DownloadURL downloads a .agent artifact from a full URL (no base + file join).
func DownloadURL(u string) (string, error) {
    req, _ := http.NewRequest("GET", u, nil)
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return "", fmt.Errorf("registry download status %d", resp.StatusCode) }
    name := filepath.Base(resp.Request.URL.Path)
    if !strings.HasSuffix(name, ".agent") { name = name + ".agent" }
    os.MkdirAll(".downloads", 0o755)
    out := filepath.Join(".downloads", name)
    f, err := os.Create(out)
    if err != nil { return "", err }
    defer f.Close()
    if _, err := io.Copy(f, resp.Body); err != nil { return "", err }
    return out, nil
}

func fetch(u string) ([]byte, error) {
    req, _ := http.NewRequest("GET", u, nil)
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return nil, fmt.Errorf("status %d", resp.StatusCode) }
    return io.ReadAll(resp.Body)
}

func verifySHA256(path, expected string) error {
    b, err := os.ReadFile(path)
    if err != nil { return err }
    h := sha256.Sum256(b)
    sum := fmt.Sprintf("%x", h[:])
    if !strings.EqualFold(sum, expected) {
        return fmt.Errorf("expected %s got %s", expected, sum)
    }
    return nil
}
