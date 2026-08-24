package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDsListCmdSearchFlag(t *testing.T) {
	cmd := newIDsListCmd()

	flag := cmd.Flags().Lookup("search")
	require.NotNil(t, flag)
	assert.Equal(t, "fuzzy search by EdgeWorker ID or name", flag.Usage)
	assert.Error(t, cmd.Args(cmd, []string{"123"}))
}
