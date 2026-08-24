package devserver

import "testing"

func TestNewRootCmd_RegistersSubcommands(t *testing.T) {
	cmd := NewRootCmd("1.2.3")
	if cmd.Version != "1.2.3" {
		t.Fatalf("unexpected version: %s", cmd.Version)
	}
	want := []string{"run", "serve", "playground"}
	for _, use := range want {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == use {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected subcommand %q to be registered", use)
		}
	}
}
