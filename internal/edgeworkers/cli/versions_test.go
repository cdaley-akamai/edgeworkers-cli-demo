package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBundleDirectory(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "main.js"), []byte("export function onClientRequest() {}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "bundle.json"), []byte(`{"edgeworker-version":"release-2026.08"}`), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(directory, "lib"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "lib", "helper.js"), []byte("export const answer = 42;\n"), 0o644))

	bundle, err := bundleDirectory(directory)
	require.NoError(t, err)

	gzipReader, err := gzip.NewReader(bytes.NewReader(bundle))
	require.NoError(t, err)
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	files := map[string]string{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Zero(t, header.Mode&0o111)
		contents, err := io.ReadAll(tarReader)
		require.NoError(t, err)
		files[header.Name] = string(contents)
	}

	require.Equal(t, map[string]string{
		"bundle.json":   `{"edgeworker-version":"release-2026.08"}`,
		"lib/helper.js": "export const answer = 42;\n",
		"main.js":       "export function onClientRequest() {}\n",
	}, files)
}

func TestBundleDirectoryRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, directory string)
		want  string
	}{
		{
			name: "missing entry point",
			setup: func(t *testing.T, directory string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(directory, "bundle.json"), []byte(`{"edgeworker-version":"release-2026.08"}`), 0o644))
			},
			want: "bundle must contain main.js",
		},
		{
			name: "schema violation",
			setup: func(t *testing.T, directory string) {
				t.Helper()
				writeBundleFiles(t, directory, []byte(`{"edgeworker-version":"release..2026"}`))
			},
			want: "validate bundle.json",
		},
		{
			name: "executable file",
			setup: func(t *testing.T, directory string) {
				t.Helper()
				writeBundleFiles(t, directory, []byte(`{"edgeworker-version":"release-2026.08"}`))
				executablePath := filepath.Join(directory, "build.sh")
				require.NoError(t, os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o755))
			},
			want: "bundle contains executable file build.sh",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			test.setup(t, directory)

			_, err := bundleDirectory(directory)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestBundleManifestSchema(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		valid    bool
	}{
		{
			name:     "valid manifest",
			manifest: `{"edgeworker-version":"release-2026.08","bundle-version":1,"api-version":"1.2","config":{"logging":{"level":"info","schema":"v1"}}}`,
			valid:    true,
		},
		{
			name:     "missing required version",
			manifest: `{}`,
		},
		{
			name:     "consecutive periods in version",
			manifest: `{"edgeworker-version":"release..2026"}`,
		},
		{
			name:     "single period version",
			manifest: `{"edgeworker-version":"."}`,
		},
		{
			name:     "unknown top level property",
			manifest: `{"edgeworker-version":"release-2026","unknown":true}`,
		},
		{
			name:     "invalid log level",
			manifest: `{"edgeworker-version":"release-2026","config":{"logging":{"level":"verbose"}}}`,
		},
		{
			name:     "mixed case log level",
			manifest: `{"edgeworker-version":"release-2026","config":{"logging":{"level":"INFO"}}}`,
		},
		{
			name:     "subrequest missing required property",
			manifest: `{"edgeworker-version":"release-2026","config":{"subrequest":{}}}`,
		},
		{
			name:     "dependency missing version",
			manifest: `{"edgeworker-version":"release-2026","dependencies":{"other":{"edgeWorkerId":123}}}`,
		},
		{
			name:     "invalid dependency version",
			manifest: `{"edgeworker-version":"release-2026","dependencies":{"other":{"edgeWorkerId":123,"version":"latest"}}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateBundleManifest([]byte(test.manifest))
			if test.valid {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
		})
	}
}

func writeBundleFiles(t *testing.T, directory string, manifest []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "main.js"), []byte("export function onClientRequest() {}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "bundle.json"), manifest, 0o644))
}
