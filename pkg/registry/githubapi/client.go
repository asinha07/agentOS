package githubapi

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

var ghRefRe = regexp.MustCompile(`^(?:github\.com/)?([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)(?:@([A-Za-z0-9_.\-/]+))?$`)

type release struct{
    TagName string `json:"tag_name"`
    Assets []struct{
        Name string `json:"name"`
        BrowserDownloadURL string `json:"browser_download_url"`
    } `json:"assets"`
}

func parseRef(ref string) (owner, repo, tag string, err error) {
    m := ghRefRe.FindStringSubmatch(ref)
    if m == nil { return "","","", fmt.Errorf("invalid github ref: %s", ref) }
    owner, repo, tag = m[1], m[2], m[3]
    return
}

func ghGet(url string) (*http.Response, error) {
    req, _ := http.NewRequest("GET", url, nil)
    if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
        req.Header.Set("Authorization", "Bearer "+tok)
    }
    req.Header.Set("Accept", "application/vnd.github+json")
    client := &http.Client{}
    return client.Do(req)
}

// DownloadAgentAsset downloads the first .agent asset from a release and returns the path.
func DownloadAgentAsset(owner, repo, tag string) (string, error) {
    var url string
    if tag == "" {
        url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
    } else {
        url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
    }
    resp, err := ghGet(url)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return "", fmt.Errorf("github api status %d", resp.StatusCode) }
    var rel release
    if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil { return "", err }
    var assetURL, name string
    for _, a := range rel.Assets {
        if strings.HasSuffix(a.Name, ".agent") { assetURL = a.BrowserDownloadURL; name = a.Name; break }
    }
    if assetURL == "" { return "", errors.New("no .agent asset found in release") }
    // download asset
    ar, err := ghGet(assetURL)
    if err != nil { return "", err }
    defer ar.Body.Close()
    if ar.StatusCode != 200 { return "", fmt.Errorf("asset download status %d", ar.StatusCode) }
    _ = os.MkdirAll(".downloads", 0o755)
    out := filepath.Join(".downloads", name)
    f, err := os.Create(out)
    if err != nil { return "", err }
    defer f.Close()
    if _, err := io.Copy(f, ar.Body); err != nil { return "", err }
    return out, nil
}

// InstallRef downloads and returns local path to .agent for a GitHub ref like owner/repo[@tag].
func InstallRef(ref string) (string, error) {
    owner, repo, tag, err := parseRef(ref)
    if err != nil { return "", err }
    return DownloadAgentAsset(owner, repo, tag)
}

