package githubapi

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "os"
)

type Repo struct{
    FullName string `json:"full_name"`
    HTMLURL string `json:"html_url"`
    Description string `json:"description"`
}

func SearchRepos(q string) ([]Repo, error) {
    base := "https://api.github.com/search/repositories"
    qs := url.Values{}
    // Search repos tagged with the topic 'agentos-agent' plus user query
    qs.Set("q", fmt.Sprintf("topic:agentos-agent %s", q))
    req, _ := http.NewRequest("GET", base+"?"+qs.Encode(), nil)
    if tok := os.Getenv("GITHUB_TOKEN"); tok != "" { req.Header.Set("Authorization", "Bearer "+tok) }
    req.Header.Set("Accept", "application/vnd.github+json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return nil, fmt.Errorf("github search status %d", resp.StatusCode) }
    var out struct{ Items []Repo `json:"items"` }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
    return out.Items, nil
}

