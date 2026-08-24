package devmanager

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManifest_FindRelease(t *testing.T) {
	manifest := Manifest{
		Current:  Release{Version: "1.1.0"},
		Previous: []Release{{Version: "1.0.0"}},
	}

	if _, err := manifest.FindRelease("1.1.0"); err != nil {
		t.Fatalf("expected current release to be found: %v", err)
	}
	if _, err := manifest.FindRelease("1.0.0"); err != nil {
		t.Fatalf("expected previous release to be found: %v", err)
	}
	if _, err := manifest.FindRelease("9.9.9"); err == nil {
		t.Fatal("expected error for unknown version")
	}
}

func TestManager_Manifest_FetchesAndParses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"current":{"version":"2.0.0","releaseDate":"2026-01-01","packages":[]},"previous":[]}`))
	}))
	defer server.Close()

	manager := &Manager{Root: t.TempDir(), ManifestURL: server.URL, httpClient: http.DefaultClient}
	manifest, err := manager.Manifest(t.Context())
	if err != nil {
		t.Fatalf("Manifest returned error: %v", err)
	}
	if manifest.Current.Version != "2.0.0" {
		t.Fatalf("unexpected current version: %s", manifest.Current.Version)
	}
}

func TestManifestURL_EnvOverride(t *testing.T) {
	t.Setenv(ManifestURLEnvVar, "https://example.com/manifest.json")
	if got := ManifestURL(); got != "https://example.com/manifest.json" {
		t.Fatalf("expected override URL, got %s", got)
	}
}

func TestManifestURL_Default(t *testing.T) {
	t.Setenv(ManifestURLEnvVar, "")
	if got := ManifestURL(); got != DefaultManifestURL() {
		t.Fatalf("expected default URL, got %s", got)
	}
}
