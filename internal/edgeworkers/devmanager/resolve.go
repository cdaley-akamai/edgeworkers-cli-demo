package devmanager

import (
	"fmt"
	"os"
)

// ResolveBinary returns the installed EdgeWorkersDevServer binary path for version.
// If version is empty, the locally recorded current version is used.
func (m *Manager) ResolveBinary(version string) (string, error) {
	if version == "" {
		state, err := m.State()
		if err != nil {
			return "", err
		}
		if state.Current == "" {
			return "", fmt.Errorf("no EdgeWorkersDevServer version installed; run 'dev update' or 'dev install <version>'")
		}
		version = state.Current
	}
	path := m.binaryPath(version)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("EdgeWorkersDevServer version %s is not installed; run 'dev install %s'", version, version)
		}
		return "", fmt.Errorf("stat binary: %w", err)
	}
	return path, nil
}
