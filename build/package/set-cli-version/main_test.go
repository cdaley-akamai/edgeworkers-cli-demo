package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleCLI = `{
  "name": "akamai/edgeworkers",
  "requirements": {
    "go": "1.26.0"
  },
  "commands": [
    {
      "name": "edgeworkers",
      "version": "2.0.0",
      "aliases": ["ew", "edgeworkers"],
      "description": "Manage Akamai EdgeWorkers.",
      "bin": "https://example.com/edgeworkers-{{.Version}}"
    },
    {
      "name": "edgekv",
      "version": "2.0.0",
      "aliases": ["ekv", "edgekv"],
      "description": "Manage Akamai EdgeKV.",
      "bin": "https://example.com/edgekv-{{.Version}}"
    }
  ]
}
`

func writeSampleCLI(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cli.json")
	if err := os.WriteFile(path, []byte(sampleCLI), 0o644); err != nil {
		t.Fatalf("write sample cli.json: %v", err)
	}
	return path
}

func TestRun_UpdatesOnlyNamedCommand(t *testing.T) {
	path := writeSampleCLI(t)

	if err := run(path, "edgeworkers", "2.1.0"); err != nil {
		t.Fatalf("run: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	got := string(data)

	if !strings.Contains(got, `"name": "edgeworkers",
      "version": "2.1.0"`) {
		t.Fatalf("expected edgeworkers version to be updated, got:\n%s", got)
	}
	if !strings.Contains(got, `"name": "edgekv",
      "version": "2.0.0"`) {
		t.Fatalf("expected edgekv version to be left untouched, got:\n%s", got)
	}
	if !strings.Contains(got, `"aliases": ["ew", "edgeworkers"],`) {
		t.Fatalf("expected unrelated formatting to be preserved, got:\n%s", got)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	path := writeSampleCLI(t)

	if err := run(path, "bogus", "1.0.0"); err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestRun_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")

	if err := run(path, "edgeworkers", "1.0.0"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
