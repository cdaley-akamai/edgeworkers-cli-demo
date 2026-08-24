package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// sliceSchemaOverrides forces slice-typed fields to a plain "array" schema.
// Unlike pointer fields, jsonschema-go always widens Go slices to ["null",
// "array"] because a nil slice is a distinct zero value; MCP clients treat
// that union as "field may be omitted", which the omitempty tag already
// conveys, so the null branch only adds noise.
var sliceSchemaOverrides = &jsonschema.ForOptions{
	TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeOf([]string(nil)): {
			Type:  "array",
			Items: &jsonschema.Schema{Type: "string"},
		},
	},
}

// Config defines product-neutral MCP server identity and runtime behavior.
type Config struct {
	// Name is the programmatic MCP implementation name.
	Name string
	// Version is the enclosing CLI release version.
	Version string
	// ProductName is the human-readable product-specific server name.
	ProductName string
	// ToolPatterns contains case-sensitive globs selecting product tools.
	ToolPatterns []string
	// LogOutput receives operational SDK logs. It defaults to stderr.
	LogOutput io.Writer
	// Debug enables debug-level operational logs.
	Debug bool
	// Now supplies hello_world timestamps. It defaults to time.Now.
	Now func() time.Time
}

// Server wraps one official SDK server with shared filtering and result behavior.
type Server struct {
	server *mcp.Server
	filter toolFilter
}

type emptyInput struct{}

// New creates a product MCP server and always registers hello_world.
func New(config Config) (*Server, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("MCP server name is required")
	}
	if config.Version == "" {
		return nil, fmt.Errorf("MCP server version is required")
	}
	if config.ProductName == "" {
		return nil, fmt.Errorf("MCP product name is required")
	}
	filter, err := newToolFilter(config.ToolPatterns)
	if err != nil {
		return nil, err
	}
	if config.LogOutput == nil {
		config.LogOutput = os.Stderr
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	level := slog.LevelError
	if config.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(config.LogOutput, &slog.HandlerOptions{Level: level}))
	server := mcp.NewServer(
		&mcp.Implementation{Name: config.Name, Title: config.ProductName, Version: config.Version},
		&mcp.ServerOptions{
			Logger: logger,
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{},
			},
		},
	)
	runtime := &Server{server: server, filter: filter}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hello_world",
		Description: fmt.Sprintf("Verify the %s server is running", config.ProductName),
	}, func(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, any, error) {
		payload, err := json.Marshal(map[string]string{
			"message":   "Hello from " + config.ProductName,
			"timestamp": config.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		})
		if err != nil {
			return nil, nil, err
		}
		return TextResult(string(payload)), nil, nil
	})
	return runtime, nil
}

// ToolHandler handles one validated, typed tool call.
type ToolHandler[Input any] func(context.Context, Input) (*mcp.CallToolResult, error)

// AddTool registers a typed product tool when its name matches the configured filters.
func AddTool[Input any](runtime *Server, tool *mcp.Tool, handler ToolHandler[Input]) {
	if !runtime.filter.matches(tool.Name) {
		return
	}
	if tool.InputSchema == nil {
		schema, err := jsonschema.For[Input](sliceSchemaOverrides)
		if err != nil {
			panic(fmt.Sprintf("mcpserver: inferring input schema for %q: %v", tool.Name, err))
		}
		tool.InputSchema = schema
	}
	mcp.AddTool(runtime.server, tool, func(ctx context.Context, _ *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, any, error) {
		result, err := handler(ctx, input)
		return result, nil, err
	})
}

// Connect starts one MCP session over transport.
func (s *Server) Connect(ctx context.Context, transport mcp.Transport) (*mcp.ServerSession, error) {
	return s.server.Connect(ctx, transport, nil)
}

// Run serves one MCP session until its transport closes or ctx is cancelled.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.server.Run(ctx, transport)
}

// RunStdio serves one MCP session over stdin and stdout.
func (s *Server) RunStdio(ctx context.Context) error {
	stdioCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := s.Run(stdioCtx, &mcp.StdioTransport{})
	if err != nil && stdioCtx.Err() != nil && ctx.Err() == nil {
		return nil
	}
	return err
}
