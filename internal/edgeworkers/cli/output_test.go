package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderJSON_Stdout(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	err := renderJSON(&buf, data)
	require.NoError(t, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "value", got["key"])
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"ID", "NAME"}
	rows := [][]string{{"123", "test-ew"}, {"456", "other-ew"}}
	renderTable(&buf, headers, rows)
	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "test-ew")
	assert.Contains(t, out, "456")
}

func TestFormatError(t *testing.T) {
	var buf bytes.Buffer
	printAPIError(&buf, "activate EdgeWorker 123", &edgeworkersapi.APIError{
		StatusCode: 403,
		Detail:     "Permission denied.",
		Instance:   "req-abc",
	}, "Check your .edgerc capabilities.")
	out := buf.String()
	assert.Contains(t, out, "activate EdgeWorker 123")
	assert.Contains(t, out, "403")
	assert.Contains(t, out, "Permission denied.")
	assert.Contains(t, out, "Check your .edgerc capabilities.")
	assert.Contains(t, out, "req-abc")
}
