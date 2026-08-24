package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/akamai/edgeworkers-cli/internal/edgeworkers/devmanager"
	"github.com/stretchr/testify/require"
)

func TestDevCommandValidatesArguments(t *testing.T) {
	command := newDevCmd()
	for _, args := range [][]string{
		{"run", "extra"},
		{"serve", "extra"},
		{"playground", "extra"},
		{"update", "extra"},
		{"install"},
		{"install", "1.2.3", "extra"},
		{"list", "extra"},
	} {
		command.SetArgs(args)
		err := command.Execute()
		require.Error(t, err)
	}
}

func TestDevCommandExposesOnlyApprovedLeaves(t *testing.T) {
	command := newDevCmd()
	names := make([]string, 0, len(command.Commands()))
	for _, child := range command.Commands() {
		names = append(names, child.Name())
	}
	require.Equal(t, []string{"install", "list", "playground", "run", "serve", "update"}, names)
}

func TestDevRunCmd_ErrorsWhenNotInstalled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	command := newDevCmd()
	command.SetArgs([]string{"run", "--server-version", "1.2.3"})
	err := command.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not installed")
}

func TestDevServeCmd_ErrorsWhenNotInstalled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	command := newDevCmd()
	command.SetArgs([]string{"serve", "--server-version", "1.2.3"})
	err := command.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not installed")
}

func TestDevPlaygroundCmd_ErrorsWhenNotInstalled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	command := newDevCmd()
	command.SetArgs([]string{"playground", "--server-version", "1.2.3"})
	err := command.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not installed")
}

func TestDevInstallCmd_InstallsVersionFromManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manifestServer := newTestManifestServer(t, "1.0.0")
	t.Setenv(devmanager.ManifestURLEnvVar, manifestServer.URL)

	command := newDevCmd()
	var out bytes.Buffer
	command.SetOut(&out)
	command.SetArgs([]string{"install", "1.0.0"})
	require.NoError(t, command.Execute())
	require.Contains(t, out.String(), "Installed EdgeWorkersDevServer 1.0.0")
}

func TestDevInstallCmd_UnknownVersion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manifestServer := newTestManifestServer(t, "1.0.0")
	t.Setenv(devmanager.ManifestURLEnvVar, manifestServer.URL)

	command := newDevCmd()
	command.SetArgs([]string{"install", "9.9.9"})
	err := command.Execute()
	require.Error(t, err)
}

func TestDevUpdateCmd_InstallsWhenOutdated(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manifestServer := newTestManifestServer(t, "1.0.0")
	t.Setenv(devmanager.ManifestURLEnvVar, manifestServer.URL)

	command := newDevCmd()
	var out bytes.Buffer
	command.SetOut(&out)
	command.SetArgs([]string{"update"})
	require.NoError(t, command.Execute())
	require.Contains(t, out.String(), "Updated EdgeWorkersDevServer to version 1.0.0")
}

func TestDevUpdateCmd_AlreadyUpToDate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manifestServer := newTestManifestServer(t, "1.0.0")
	t.Setenv(devmanager.ManifestURLEnvVar, manifestServer.URL)

	first := newDevCmd()
	first.SetArgs([]string{"update"})
	require.NoError(t, first.Execute())

	second := newDevCmd()
	var out bytes.Buffer
	second.SetOut(&out)
	second.SetArgs([]string{"update"})
	require.NoError(t, second.Execute())
	require.Contains(t, out.String(), "already up to date")
}

func TestDevListCmd_RendersVersionTable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manifestServer := newTestManifestServer(t, "1.0.0")
	t.Setenv(devmanager.ManifestURLEnvVar, manifestServer.URL)

	command := newDevCmd()
	command.SetArgs([]string{"list"})
	stdout := captureStdout(t, func() {
		require.NoError(t, command.Execute())
	})
	require.Contains(t, stdout, "1.0.0")
	require.Contains(t, stdout, "yes")
}

// newTestManifestServer serves a manifest whose current release downloads from an embedded
// package server, matching the running platform's OS/arch.
func newTestManifestServer(t *testing.T, version string) *httptest.Server {
	t.Helper()
	content := []byte("fake-devserver-binary")
	sum := sha256.Sum256(content)
	checksum := hex.EncodeToString(sum[:])

	pkgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(content)
	}))
	t.Cleanup(pkgServer.Close)

	manifest := fmt.Sprintf(
		`{"current":{"version":%q,"releaseDate":"2026-01-01","packages":[{"os":%q,"arch":%q,"url":%q,"sha256":%q}]},"previous":[]}`,
		version, runtime.GOOS, runtime.GOARCH, pkgServer.URL, checksum,
	)
	manifestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(manifestServer.Close)
	return manifestServer
}

// captureStdout redirects os.Stdout for the duration of fn and returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = original }()

	fn()

	require.NoError(t, w.Close())
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data)
}
