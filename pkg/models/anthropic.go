package models

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "time"
)

// Anthropic adapter (Messages API)
// Env: ANTHROPIC_API_KEY; Header: anthropic-version: 2023-06-01
type Anthropic struct{ Model string }

func (a Anthropic) Generate(prompt string) (string, error) {
    key := os.Getenv("ANTHROPIC_API_KEY")
    if key == "" { return "", errors.New("ANTHROPIC_API_KEY not set") }
    if a.Model == "" { a.Model = "claude-3-5-sonnet-latest" }
    body := map[string]any{
        "model": a.Model,
        "max_tokens": 512,
        "messages": []map[string]any{{"role": "user", "content": prompt}},
    }
    b, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(b))
    req.Header.Set("x-api-key", key)
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("content-type", "application/json")
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return "", fmt.Errorf("anthropic status %d", resp.StatusCode)
    }
    var out struct{
        Content []struct{ Text string `json:"text"` } `json:"content"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "", err }
    if len(out.Content) == 0 { return "", errors.New("no content") }
    return out.Content[0].Text, nil
}

func (a Anthropic) Stream(prompt string, onToken func(tok string)) error {
    s, err := a.Generate(prompt)
    if err != nil { return err }
    onToken(s)
    return nil
}

