package models

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"
)

// OpenAI makes simple non-streaming requests to Chat Completions API.
type OpenAI struct{ Model string }

func (o OpenAI) Generate(prompt string) (string, error) {
    key := os.Getenv("OPENAI_API_KEY")
    if key == "" { return "", errors.New("OPENAI_API_KEY not set") }
    if o.Model == "" { o.Model = "gpt-4o-mini" }
    // Try Chat Completions first for broad compatibility
    content, status, err := o.generateChatCompletions(key, prompt)
    if err == nil { return content, nil }
    // If model is gpt-4.1 or chat failed with client error, try Responses API
    if strings.HasPrefix(o.Model, "gpt-4.1") || (status >= 400 && status < 500) {
        if content2, err2 := o.generateResponses(key, prompt); err2 == nil {
            return content2, nil
        }
    }
    return "", err
}

func (o OpenAI) Stream(prompt string, onToken func(tok string)) error {
    // Streaming not implemented in prototype.
    s, err := o.Generate(prompt)
    if err != nil { return err }
    onToken(s)
    return nil
}

func (o OpenAI) generateChatCompletions(key, prompt string) (string, int, error) {
    body := map[string]any{
        "model": o.Model,
        "messages": []map[string]string{{"role": "user", "content": prompt}},
        "temperature": 0.3,
    }
    b, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+key)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return "", 0, err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return "", resp.StatusCode, fmt.Errorf("openai status %d", resp.StatusCode)
    }
    var out struct{ Choices []struct{ Message struct{ Content string `json:"content"` } `json:"message"` } `json:"choices"` }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "", resp.StatusCode, err }
    if len(out.Choices) == 0 { return "", resp.StatusCode, errors.New("no choices") }
    return out.Choices[0].Message.Content, resp.StatusCode, nil
}

func (o OpenAI) generateResponses(key, prompt string) (string, error) {
    body := map[string]any{
        "model": o.Model,
        "input": prompt,
    }
    b, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+key)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return "", fmt.Errorf("openai responses status %d", resp.StatusCode)
    }
    // Parse generically to extract text
    var raw map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil { return "", err }
    // Try output_text (array of strings)
    if ot, ok := raw["output_text"].([]any); ok && len(ot) > 0 {
        var s strings.Builder
        for _, v := range ot { if str, ok := v.(string); ok { s.WriteString(str) } }
        if s.Len() > 0 { return s.String(), nil }
    }
    // Try output -> [] -> content -> [] -> text
    if out, ok := raw["output"].([]any); ok && len(out) > 0 {
        var s strings.Builder
        for _, item := range out {
            if m, ok := item.(map[string]any); ok {
                if content, ok := m["content"].([]any); ok {
                    for _, seg := range content {
                        if sm, ok := seg.(map[string]any); ok {
                            if txt, ok := sm["text"].(string); ok { s.WriteString(txt) }
                        }
                    }
                }
            }
        }
        if s.Len() > 0 { return s.String(), nil }
    }
    // Fallback to raw string
    return "", errors.New("unable to parse responses payload")
}
