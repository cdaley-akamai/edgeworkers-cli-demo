package mcpserver

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type testToolInput struct {
	Name string `json:"name"`
}

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

func TestServerInitializesWithFilteredToolsOnly(t *testing.T) {
	runtime, err := New(Config{
		Name:         "akamai-edgeworkers-mcp",
		Version:      "9.8.7",
		ProductName:  "Akamai EdgeWorkers MCP",
		ToolPatterns: []string{"list*"},
		LogOutput:    io.Discard,
		Now: func() time.Time {
			return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
		},
	})
	require.NoError(t, err)

	AddTool(runtime, &mcp.Tool{Name: "listThings", Description: "List things."}, func(_ context.Context, input testToolInput) (*mcp.CallToolResult, error) {
		return JSONResult(map[string]string{"name": input.Name})
	})
	AddTool(runtime, &mcp.Tool{Name: "deleteThing", Description: "Delete a thing."}, func(context.Context, testToolInput) (*mcp.CallToolResult, error) {
		return JSONResult(map[string]bool{"deleted": true})
	})

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := runtime.Connect(ctx, serverTransport)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, &mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}})
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	initialization := clientSession.InitializeResult()
	require.Equal(t, "akamai-edgeworkers-mcp", initialization.ServerInfo.Name)
	require.Equal(t, "9.8.7", initialization.ServerInfo.Version)
	require.NotNil(t, initialization.Capabilities.Tools)
	require.False(t, initialization.Capabilities.Tools.ListChanged)
	require.Nil(t, initialization.Capabilities.Logging)
	require.Nil(t, initialization.Capabilities.Prompts)
	require.Nil(t, initialization.Capabilities.Resources)

	tools, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 2)
	require.Equal(t, "hello_world", tools.Tools[0].Name)
	require.Equal(t, "listThings", tools.Tools[1].Name)

	hello, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "hello_world"})
	require.NoError(t, err)
	require.False(t, hello.IsError)
	require.Equal(t, `{"message":"Hello from Akamai EdgeWorkers MCP","timestamp":"2026-01-02T03:04:05.000Z"}`, hello.Content[0].(*mcp.TextContent).Text)
}

func TestServerConvertsHandlerErrorsToToolErrors(t *testing.T) {
	runtime, err := New(Config{
		Name:         "test-server",
		Version:      "dev",
		ProductName:  "Test MCP",
		ToolPatterns: []string{"*"},
		LogOutput:    io.Discard,
	})
	require.NoError(t, err)

	AddTool(runtime, &mcp.Tool{Name: "fails"}, func(context.Context, testToolInput) (*mcp.CallToolResult, error) {
		return nil, io.ErrUnexpectedEOF
	})

	clientSession := connectTestClient(t, runtime)
	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{Name: "fails", Arguments: map[string]any{"name": "value"}})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Equal(t, "unexpected EOF", result.Content[0].(*mcp.TextContent).Text)
}

func TestServerUsesSDKInputValidationBeforeCallingHandler(t *testing.T) {
	runtime, err := New(Config{
		Name:         "test-server",
		Version:      "dev",
		ProductName:  "Test MCP",
		ToolPatterns: []string{"*"},
		LogOutput:    io.Discard,
	})
	require.NoError(t, err)

	handlerCalls := 0
	AddTool(runtime, &mcp.Tool{Name: "echo"}, func(context.Context, testToolInput) (*mcp.CallToolResult, error) {
		handlerCalls++
		return TextResult("called"), nil
	})

	clientSession := connectTestClient(t, runtime)
	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"name": "value", "unexpected": true},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Zero(t, handlerCalls)
	require.Contains(t, strings.ToLower(result.Content[0].(*mcp.TextContent).Text), "additional")
}

func TestRunStdioServesProtocolFrames(t *testing.T) {
	command, clientSession, stderr := startStdioHelper(t)

	tools, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 1)
	require.Equal(t, "hello_world", tools.Tools[0].Name)
	require.NoError(t, clientSession.Close())
	require.NoErrorf(t, waitForStdioHelper(t, command), "helper stderr: %s", stderr.String())
}

func TestRunStdioStopsCleanlyOnSignals(t *testing.T) {
	for _, shutdownSignal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(shutdownSignal.String(), func(t *testing.T) {
			command, clientSession, stderr := startStdioHelper(t)

			result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{Name: "hello_world"})
			require.NoError(t, err)
			require.False(t, result.IsError)
			require.NoError(t, command.Process.Signal(shutdownSignal))
			require.NoErrorf(t, waitForStdioHelper(t, command), "helper stderr: %s", stderr.String())
			_ = clientSession.Close()
		})
	}
}

func TestRunReturnsCallerCancellation(t *testing.T) {
	runtime, err := New(Config{
		Name:         "cancellation-test-server",
		Version:      "1.0.0",
		ProductName:  "Cancellation Test MCP",
		ToolPatterns: []string{"*"},
		LogOutput:    io.Discard,
	})
	require.NoError(t, err)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runtime.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(
		&mcp.Implementation{Name: "cancellation-test-client", Version: "1.0.0"},
		&mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}},
	)
	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err)
	cancel()
	require.ErrorIs(t, <-done, context.Canceled)
	_ = clientSession.Close()
}

func TestRunStdioHelper(t *testing.T) {
	if os.Getenv("MCP_STDIO_HELPER") != "1" {
		t.Skip("subprocess helper")
	}
	runtime, err := New(Config{
		Name:         "stdio-test-server",
		Version:      "1.0.0",
		ProductName:  "Stdio Test MCP",
		ToolPatterns: []string{"*"},
		LogOutput:    io.Discard,
	})
	if err != nil {
		os.Exit(2)
	}
	if err := runtime.RunStdio(context.Background()); err != nil {
		os.Exit(3)
	}
	os.Exit(0)
}

func startStdioHelper(t *testing.T) (*exec.Cmd, *mcp.ClientSession, *synchronizedBuffer) {
	t.Helper()
	command := exec.Command(os.Args[0], "-test.run=^TestRunStdioHelper$")
	command.Env = append(os.Environ(), "MCP_STDIO_HELPER=1")
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
		&mcp.Implementation{Name: "stdio-test-client", Version: "1.0.0"},
		&mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}},
	)
	clientSession, err := client.Connect(context.Background(), &mcp.IOTransport{Reader: stdout, Writer: stdin}, nil)
	require.NoErrorf(t, err, "helper stderr: %s", stderr.String())
	return command, clientSession, stderr
}

func waitForStdioHelper(t *testing.T, command *exec.Cmd) error {
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
		t.Fatal("stdio helper did not exit")
		return nil
	}
}

func connectTestClient(t *testing.T, runtime *Server) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := runtime.Connect(ctx, serverTransport)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, &mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}})
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })
	return clientSession
}
