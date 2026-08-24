package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/pflag"
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

func TestJSONFlagDefaultsToStdout(t *testing.T) {
	var jsonFlag bool
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.BoolVar(&jsonFlag, "json", false, "")

	require.NoError(t, flags.Parse([]string{"--json"}))
	assert.True(t, jsonFlag)
}

func TestJSONFlagIsFalseByDefault(t *testing.T) {
	var jsonFlag bool
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.BoolVar(&jsonFlag, "json", false, "")

	require.NoError(t, flags.Parse(nil))
	assert.False(t, jsonFlag)
}
