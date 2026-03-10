package oci

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "path/filepath"

    "github.com/google/go-containerregistry/pkg/name"
    "github.com/google/go-containerregistry/pkg/v1/empty"
    "github.com/google/go-containerregistry/pkg/v1/mutate"
    "github.com/google/go-containerregistry/pkg/v1/remote"
    "github.com/google/go-containerregistry/pkg/v1/static"
    crtypes "github.com/google/go-containerregistry/pkg/v1/types"
)

// PushLib uploads the .agent artifact as a single-layer OCI image using go-containerregistry.
func PushLib(ref, artifactPath string) error {
    b, err := os.ReadFile(artifactPath)
    if err != nil { return err }
    layer := static.NewLayer(b, crtypes.MediaType(MediaTypeAgent))
    img := empty.Image
    img, err = mutate.AppendLayers(img, layer)
    if err != nil { return err }
    // Note: Manifest annotations omitted in prototype to ensure compatibility across versions
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
    // Note: No manifest annotation verification in prototype path
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
