package oci

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

const MediaTypeAgent = "application/vnd.agentos.agent"

// Push uses the external ORAS CLI to upload a .agent artifact to an OCI registry reference
// (e.g., ghcr.io/org/agent:1.0.0). Requires the `oras` binary.
func Push(ref string, artifactPath string) error {
    // Try library first
    if err := PushLib(ref, artifactPath); err == nil {
        return nil
    }
    // Fallback to ORAS CLI
    if _, err := exec.LookPath("oras"); err != nil {
        return fmt.Errorf("oras CLI not found and library push failed")
    }
    cmd := exec.Command("oras", "push", ref, fmt.Sprintf("%s:%s", artifactPath, MediaTypeAgent))
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

// Pull uses the external ORAS CLI to download a .agent artifact from an OCI reference.
// Returns the downloaded file path under .downloads/.
func Pull(ref string) (string, error) {
    if p, err := PullLib(ref); err == nil {
        return p, nil
    }
    if _, err := exec.LookPath("oras"); err != nil {
        return "", fmt.Errorf("oras CLI not found and library pull failed")
    }
    outDir := ".downloads"
    _ = os.MkdirAll(outDir, 0o755)
    cmd := exec.Command("oras", "pull", ref)
    cmd.Dir = outDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return "", err
    }
    entries, _ := os.ReadDir(outDir)
    for _, e := range entries {
        if e.IsDir() { continue }
        if filepath.Ext(e.Name()) == ".agent" {
            return filepath.Join(outDir, e.Name()), nil
        }
    }
    return "", fmt.Errorf("no .agent file found after oras pull")
}
