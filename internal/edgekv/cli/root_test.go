package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommandReportsVersion(t *testing.T) {
	cmd := NewRootCmd("9.8.7")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--version"})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, output.String(), "9.8.7")
}

func TestRootCommandGroups(t *testing.T) {
	cmd := newRootCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--help"})

	require.NoError(t, cmd.Execute())
	for _, name := range []string{"database", "namespaces", "items", "tokens", "config"} {
		assert.Contains(t, output.String(), name)
	}
}

func TestGroupedCommandPaths(t *testing.T) {
	cmd := newRootCmd()
	for _, path := range [][]string{
		{"database", "initialize", "--help"},
		{"namespaces", "groups", "list", "--help"},
		{"namespaces", "permissions", "set", "--help"},
		{"items", "put", "--help"},
		{"tokens", "refresh", "--help"},
	} {
		cmd.SetArgs(path)
		require.NoError(t, cmd.Execute(), strings.Join(path, " "))
	}
}
