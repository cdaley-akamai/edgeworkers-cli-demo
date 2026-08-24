package devserver

import (
	"os"
	"strings"
	"testing"
)

func TestRunPipe_EchoesValidJSON(t *testing.T) {
	var out strings.Builder
	body := `{"edgeWorkerId":1234,"eventHandler":"onClientRequest","requestId":"abcd1234","resourceTier":200}`
	if err := runPipe(strings.NewReader(body), &out); err != nil {
		t.Fatalf("runPipe returned error: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != body {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunPipe_RejectsInvalidJSON(t *testing.T) {
	var out strings.Builder
	if err := runPipe(strings.NewReader(`not json`), &out); err == nil {
		t.Fatal("expected error for invalid JSON input")
	}
}

func TestIsInteractiveTerminal_CharDeviceIsInteractive(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer f.Close()
	if !isInteractiveTerminal(f) {
		t.Fatalf("expected %s to be treated as an interactive terminal", os.DevNull)
	}
}

func TestIsInteractiveTerminal_PipeIsNotInteractive(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()
	if isInteractiveTerminal(r) {
		t.Fatal("expected a pipe not to be treated as an interactive terminal")
	}
}

func TestIsInteractiveTerminal_NonFileReaderIsNotInteractive(t *testing.T) {
	if isInteractiveTerminal(strings.NewReader("")) {
		t.Fatal("expected a non-*os.File reader not to be treated as an interactive terminal")
	}
}
