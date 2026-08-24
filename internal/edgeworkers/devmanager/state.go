package devmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// State is the local record of which EdgeWorkersDevServer version is current.
type State struct {
	Current string `json:"current"`
}

func (m *Manager) statePath() string {
	return filepath.Join(m.Root, "state.json")
}

// State reads the local install state, returning a zero-value State if none exists yet.
func (m *Manager) State() (State, error) {
	data, err := os.ReadFile(m.statePath())
	if os.IsNotExist(err) {
		return State{}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("read state: %w", err)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("parse state: %w", err)
	}
	return state, nil
}

// SetCurrent records version as the current installed version.
func (m *Manager) SetCurrent(version string) error {
	if err := os.MkdirAll(m.Root, 0o755); err != nil {
		return fmt.Errorf("create install root: %w", err)
	}
	data, err := json.MarshalIndent(State{Current: version}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := os.WriteFile(m.statePath(), data, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}
