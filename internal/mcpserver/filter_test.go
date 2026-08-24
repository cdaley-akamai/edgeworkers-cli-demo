package mcpserver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToolFilterUsesPortableCaseSensitiveGlobs(t *testing.T) {
	filter, err := newToolFilter([]string{"list{Edgeworkers,Versions}", "get?roup", "create[SV]*"})
	require.NoError(t, err)

	require.True(t, filter.matches("listEdgeworkers"))
	require.True(t, filter.matches("listVersions"))
	require.True(t, filter.matches("getGroup"))
	require.True(t, filter.matches("createSecureToken"))
	require.False(t, filter.matches("ListEdgeworkers"))
	require.False(t, filter.matches("deleteEdgeworker"))
}

func TestToolFilterRejectsInvalidPattern(t *testing.T) {
	_, err := newToolFilter([]string{"list["})
	require.ErrorContains(t, err, `invalid tool pattern "list["`)
}

func TestToolFilterWithNoPatternsMatchesNothing(t *testing.T) {
	filter, err := newToolFilter(nil)
	require.NoError(t, err)
	require.False(t, filter.matches("listEdgeworkers"))
}
