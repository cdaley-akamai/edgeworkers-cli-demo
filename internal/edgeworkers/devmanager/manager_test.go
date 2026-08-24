package devmanager

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestState_DefaultsToZeroValue(t *testing.T) {
	manager := &Manager{Root: t.TempDir()}
	state, err := manager.State()
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Current != "" {
		t.Fatalf("expected empty current, got %s", state.Current)
	}
}

func TestSetCurrent_PersistsState(t *testing.T) {
	manager := &Manager{Root: t.TempDir()}
	if err := manager.SetCurrent("1.2.3"); err != nil {
		t.Fatalf("SetCurrent returned error: %v", err)
	}
	state, err := manager.State()
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Current != "1.2.3" {
		t.Fatalf("expected current 1.2.3, got %s", state.Current)
	}
}

func TestResolveBinary_ErrorsWhenNotInstalled(t *testing.T) {
	manager := &Manager{Root: t.TempDir()}
	if _, err := manager.ResolveBinary(""); err == nil {
		t.Fatal("expected error when nothing installed")
	}
}

func TestInstallAndResolveBinary(t *testing.T) {
	content := []byte("fake-binary-contents")
	sum := sha256.Sum256(content)
	checksum := hex.EncodeToString(sum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	defer server.Close()

	manager := &Manager{Root: t.TempDir(), httpClient: http.DefaultClient}
	release := Release{
		Version: "1.0.0",
		Packages: []Package{
			{OS: runtime.GOOS, Arch: runtime.GOARCH, URL: server.URL, SHA256: checksum},
		},
	}

	if err := manager.Install(t.Context(), release); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	path, err := manager.ResolveBinary("")
	if err != nil {
		t.Fatalf("ResolveBinary returned error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading installed binary: %v", err)
	}
	if string(got) != string(content) {
		t.Fatal("installed binary contents did not match downloaded package")
	}

	state, err := manager.State()
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Current != "1.0.0" {
		t.Fatalf("expected current 1.0.0, got %s", state.Current)
	}
}

func TestInstall_RejectsChecksumMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data"))
	}))
	defer server.Close()

	manager := &Manager{Root: t.TempDir(), httpClient: http.DefaultClient}
	release := Release{
		Version: "1.0.0",
		Packages: []Package{
			{OS: runtime.GOOS, Arch: runtime.GOARCH, URL: server.URL, SHA256: "deadbeef"},
		},
	}

	if err := manager.Install(t.Context(), release); err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if _, err := os.Stat(filepath.Join(manager.Root, "1.0.0")); !os.IsNotExist(err) {
		t.Fatal("expected version directory to remain absent after failed checksum")
	}
}

func TestPlatformPackage_ErrorsWhenMissing(t *testing.T) {
	release := Release{Version: "1.0.0", Packages: []Package{{OS: "plan9", Arch: "amd64"}}}
	if _, err := PlatformPackage(release); err == nil {
		t.Fatal("expected error for unmatched platform")
	}
}
