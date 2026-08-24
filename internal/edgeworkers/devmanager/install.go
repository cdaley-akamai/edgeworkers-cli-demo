package devmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

// PlatformPackage returns the release's package matching the current OS/arch.
func PlatformPackage(release Release) (Package, error) {
	for _, pkg := range release.Packages {
		if pkg.OS == runtime.GOOS && pkg.Arch == runtime.GOARCH {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("no %s/%s package available for version %s", runtime.GOOS, runtime.GOARCH, release.Version)
}

// Install downloads, verifies, and installs release for the current platform, then marks
// it as the current version.
func (m *Manager) Install(ctx context.Context, release Release) error {
	pkg, err := PlatformPackage(release)
	if err != nil {
		return err
	}
	data, err := m.download(ctx, pkg.URL)
	if err != nil {
		return err
	}
	if err := verifyChecksum(data, pkg.SHA256); err != nil {
		return err
	}
	dir := m.versionDir(release.Version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create version directory: %w", err)
	}
	path := m.binaryPath(release.Version)
	if err := os.WriteFile(path, data, 0o755); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}
	return m.SetCurrent(release.Version)
}

func (m *Manager) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build download request: %w", err)
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download package: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download package: unexpected status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read package: %w", err)
	}
	return data, nil
}

func verifyChecksum(data []byte, want string) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if want != "" && got != want {
		return fmt.Errorf("checksum mismatch: want %s, got %s", want, got)
	}
	return nil
}
