package oci

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "crypto/sha256"

    "github.com/google/go-containerregistry/pkg/name"
    v1 "github.com/google/go-containerregistry/pkg/v1"
    "github.com/google/go-containerregistry/pkg/v1/empty"
    "github.com/google/go-containerregistry/pkg/v1/mutate"
    "github.com/google/go-containerregistry/pkg/v1/remote"
    "github.com/google/go-containerregistry/pkg/v1/static"
    crtypes "github.com/google/go-containerregistry/pkg/v1/types"
    "strings"
)

// PushLib uploads the .agent artifact as a single-layer OCI image using go-containerregistry.
func PushLib(ref, artifactPath string) error {
    b, err := os.ReadFile(artifactPath)
    if err != nil { return err }
    layer := static.NewLayer(b, crtypes.MediaType(MediaTypeAgent))
    img := empty.Image
    img, err = mutate.AppendLayers(img, layer)
    if err != nil { return err }
    // Add annotations, including SHA256 of artifact
    sum := sha256.Sum256(b)
    ann := map[string]string{
        "dev.agentos.sha256": fmt.Sprintf("%x", sum[:]),
        "org.opencontainers.artifact.description": "AgentOS package",
        "org.opencontainers.artifact.type": MediaTypeAgent,
        "org.opencontainers.image.title": filepath.Base(artifactPath),
    }
    if newImg, err := mutate.Annotations(img, ann); err == nil {
        img = newImg
    }
    r, err := name.ParseReference(ref)
    if err != nil { return err }
    return remote.Write(r, img)
}

// PullLib downloads the first layer with media type MediaTypeAgent and writes it under .downloads/.
func PullLib(ref string) (string, error) {
    r, err := name.ParseReference(ref)
    if err != nil { return "", err }
    img, err := remote.Image(r)
    if err != nil { return "", err }
    layers, err := img.Layers()
    if err != nil { return "", err }
    // Verify annotations if present
    if mf, err := img.Manifest(); err == nil && mf.Annotations != nil {
        expected := mf.Annotations["dev.agentos.sha256"]
        if expected != "" {
            // Read first layer and verify hash
            if len(layers) > 0 {
                rc, _ := layers[0].Uncompressed()
                defer rc.Close()
                var buf bytes.Buffer
                io.Copy(&buf, rc)
                got := fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
                if !strings.EqualFold(got, expected) {
                    return "", fmt.Errorf("sha256 mismatch: expected %s got %s", expected, got)
                }
            }
        }
    }
    for _, l := range layers {
        mt, _ := l.MediaType()
        if string(mt) != MediaTypeAgent { continue }
        rc, err := l.Uncompressed()
        if err != nil { return "", err }
        defer rc.Close()
        var buf bytes.Buffer
        if _, err := io.Copy(&buf, rc); err != nil { return "", err }
        outDir := ".downloads"
        _ = os.MkdirAll(outDir, 0o755)
        out := filepath.Join(outDir, "pulled.agent")
        if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil { return "", err }
        return out, nil
    }
    return "", fmt.Errorf("no agent layer found")
}
