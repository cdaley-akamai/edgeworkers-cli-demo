// Package devmanager manages local installs of EdgeWorkersDevServer.
//
// It fetches a version manifest, resolves which installed version to run, and
// installs/updates versions on disk for the edgeworkers CLI's "dev" commands.
package devmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Package describes a single platform's downloadable EdgeWorkersDevServer binary.
type Package struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// Release describes one published EdgeWorkersDevServer version.
type Release struct {
	Version     string    `json:"version"`
	ReleaseDate string    `json:"releaseDate"`
	Packages    []Package `json:"packages"`
}

// Manifest is the devserver-manifest.json document listing available EdgeWorkersDevServer versions.
type Manifest struct {
	Current  Release   `json:"current"`
	Previous []Release `json:"previous"`
}

// ManifestURLEnvVar overrides the manifest location; the manifest may move off GitHub in the future.
const ManifestURLEnvVar = "AKAMAI_EDGEWORKERS_DEVSERVER_MANIFEST_URL"

const defaultManifestRepository = "cdaley-akamai/edgeworkers-cli-demo"

// DefaultManifestURL returns the manifest URL used when no override is set.
//
// It uses the refs/heads/master form rather than the plain master form because
// raw.githubusercontent.com caches each path independently, and the plain form has
// been observed to serve stale content for longer after a push.
func DefaultManifestURL() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/refs/heads/master/devserver-manifest.json", defaultManifestRepository)
}

// ManifestURL resolves the manifest URL, honoring the ManifestURLEnvVar override.
func ManifestURL() string {
	if url := os.Getenv(ManifestURLEnvVar); url != "" {
		return url
	}
	return DefaultManifestURL()
}

// FindRelease returns the release matching version among manifest.Current and manifest.Previous.
func (m *Manifest) FindRelease(version string) (Release, error) {
	if m.Current.Version == version {
		return m.Current, nil
	}
	for _, release := range m.Previous {
		if release.Version == version {
			return release, nil
		}
	}
	return Release{}, fmt.Errorf("version %q not found in manifest", version)
}

// fetchManifest downloads and parses the manifest from url using client.
func fetchManifest(ctx context.Context, client *http.Client, url string) (*Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build manifest request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch manifest: unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &manifest, nil
}
