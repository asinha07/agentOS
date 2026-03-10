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

// Grok (xAI) adapter using chat completions-like API.
// Env: XAI_API_KEY
type Grok struct{ Model string }

func (g Grok) Generate(prompt string) (string, error) {
    key := os.Getenv("XAI_API_KEY")
    if key == "" { return "", errors.New("XAI_API_KEY not set") }
    if g.Model == "" { g.Model = "grok-2" }
    body := map[string]any{
        "model": g.Model,
        "messages": []map[string]string{{"role": "user", "content": prompt}},
        "temperature": 0.3,
    }
    b, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", "https://api.x.ai/v1/chat/completions", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+key)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return "", fmt.Errorf("xai status %d", resp.StatusCode)
    }
    var out struct{ Choices []struct{ Message struct{ Content string `json:"content"` } `json:"message"` } `json:"choices"` }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "", err }
    if len(out.Choices) == 0 { return "", errors.New("no choices") }
    return out.Choices[0].Message.Content, nil
}

func (g Grok) Stream(prompt string, onToken func(tok string)) error {
    s, err := g.Generate(prompt)
    if err != nil { return err }
    onToken(s)
    return nil
}

