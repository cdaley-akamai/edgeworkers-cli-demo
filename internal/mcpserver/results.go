package mcpserver

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// JSONResult returns a successful MCP text result containing indented JSON.
func JSONResult(value any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode tool result: %w", err)
	}
	return TextResult(string(data)), nil
}

// TextResult returns a successful MCP text result.
func TextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

// AbortResult returns the normal non-error result used when confirmation is declined.
func AbortResult(operation string) *mcp.CallToolResult {
	return TextResult(fmt.Sprintf("Aborted: %s requires confirm to be true. Set confirm: true to proceed.", operation))
}
