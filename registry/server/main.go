package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type AgentInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    File    string `json:"file"`
}

func listAgents(root string) ([]AgentInfo, error) {
    var out []AgentInfo
    entries, err := os.ReadDir(root)
    if err != nil { return nil, err }
    for _, e := range entries {
        if e.IsDir() { continue }
        name := e.Name()
        if !strings.HasSuffix(name, ".agent") { continue }
        base := strings.TrimSuffix(name, ".agent")
        parts := strings.Split(base, "-")
        if len(parts) < 2 { continue }
        ver := parts[len(parts)-1]
        n := strings.Join(parts[:len(parts)-1], "-")
        out = append(out, AgentInfo{Name: n, Version: ver, File: name})
    }
    return out, nil
}

func main() {
    root := filepath.Join("registry", "agents")
    if err := os.MkdirAll(root, 0o755); err != nil {
        log.Fatalf("failed to create registry dir: %v", err)
    }
    http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
        q := r.URL.Query().Get("q")
        list, _ := listAgents(root)
        var filt []AgentInfo
        for _, a := range list {
            if q == "" || strings.Contains(a.Name, q) {
                filt = append(filt, a)
            }
        }
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(map[string]any{"agents": filt}); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    })
    fs := http.FileServer(http.Dir(root))
    http.Handle("/agents/", http.StripPrefix("/agents/", fs))
    log.Println("AgentOS Registry server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
