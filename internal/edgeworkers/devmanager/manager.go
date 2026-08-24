package devmanager

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
)

// Manager resolves, installs, and lists EdgeWorkersDevServer versions on disk.
type Manager struct {
	// Root is the directory holding per-version installs and state.json.
	Root string
	// ManifestURL is the location fetched by Manifest.
	ManifestURL string

	httpClient *http.Client
}

// NewManager creates a Manager rooted at the default install directory
// (~/.akamai/edgeworkers/devserver) using the default manifest URL resolution.
func NewManager() (*Manager, error) {
	root, err := defaultRoot()
	if err != nil {
		return nil, err
	}
	return &Manager{
		Root:        root,
		ManifestURL: ManifestURL(),
		httpClient:  http.DefaultClient,
	}, nil
}

func defaultRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".akamai", "edgeworkers", "devserver"), nil
}

// Manifest fetches and parses the version manifest from m.ManifestURL.
func (m *Manager) Manifest(ctx context.Context) (*Manifest, error) {
	return fetchManifest(ctx, m.httpClient, m.ManifestURL)
}

func (m *Manager) versionDir(version string) string {
	return filepath.Join(m.Root, version)
}

func (m *Manager) binaryPath(version string) string {
	name := "edgeworkersdevserver"
	if isWindows() {
		name += ".exe"
	}
	return filepath.Join(m.versionDir(version), name)
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
