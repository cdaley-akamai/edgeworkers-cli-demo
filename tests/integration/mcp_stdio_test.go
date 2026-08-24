//go:build !windows

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type synchronizedBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (buffer *synchronizedBuffer) Write(data []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buffer.Write(data)
}

func (buffer *synchronizedBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buffer.String()
}

func TestProductMCPCommandsServeStdioWithoutCredentials(t *testing.T) {
	testCases := []struct {
		command     string
		serverName  string
		productName string
		probeTool   string
	}{
		{command: "edgeworkers", serverName: "akamai-edgeworkers-mcp", productName: "Akamai EdgeWorkers MCP", probeTool: "listEdgeworkers"},
		{command: "edgekv", serverName: "akamai-edgekv-mcp", productName: "Akamai EdgeKV MCP", probeTool: "getEdgeKVInitializeStatus"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.command, func(t *testing.T) {
			binary := buildProductBinary(t, testCase.command)
			command := exec.Command(binary, "mcp")
			command.Env = append(os.Environ(), "HOME="+t.TempDir())
			stdin, err := command.StdinPipe()
			require.NoError(t, err)
			stdout, err := command.StdoutPipe()
			require.NoError(t, err)
			stderr := &synchronizedBuffer{}
			command.Stderr = stderr
			require.NoError(t, command.Start())
			t.Cleanup(func() {
				if command.ProcessState == nil {
					_ = command.Process.Kill()
				}
			})

			client := mcp.NewClient(
				&mcp.Implementation{Name: "integration-test-client", Version: "1.0.0"},
				&mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}},
			)
			session, err := client.Connect(context.Background(), &mcp.IOTransport{Reader: stdout, Writer: stdin}, nil)
			require.NoErrorf(t, err, "stderr: %s", stderr.String())
			assert.Equal(t, testCase.serverName, session.InitializeResult().ServerInfo.Name)

			tools, err := session.ListTools(context.Background(), nil)
			require.NoError(t, err)
			assert.NotEmpty(t, tools.Tools)
			toolNames := make([]string, 0, len(tools.Tools))
			for _, tool := range tools.Tools {
				toolNames = append(toolNames, tool.Name)
			}
			assert.Contains(t, toolNames, "hello_world")

			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "hello_world"})
			require.NoError(t, err)
			require.False(t, result.IsError)
			var hello map[string]string
			require.NoError(t, json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &hello))
			assert.Equal(t, "Hello from "+testCase.productName, hello["message"])
			assert.NotEmpty(t, hello["timestamp"])

			probe, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: testCase.probeTool})
			require.NoError(t, err)
			require.True(t, probe.IsError)
			assert.Contains(t, probe.Content[0].(*mcp.TextContent).Text, "failed to load credentials")

			require.NoError(t, command.Process.Signal(syscall.SIGTERM))
			require.NoErrorf(t, waitForCommand(t, command), "stderr: %s", stderr.String())
			_ = session.Close()
		})
	}
}

func buildProductBinary(t *testing.T, product string) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	binary := filepath.Join(t.TempDir(), product)
	command := exec.Command("go", "build", "-o", binary, "./cmd/"+product)
	command.Dir = repositoryRoot
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "build %s: %s", product, output)
	return binary
}

func waitForCommand(t *testing.T, command *exec.Cmd) error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		_ = command.Process.Kill()
		<-done
		t.Fatal("MCP command did not exit")
		return nil
	}
}
