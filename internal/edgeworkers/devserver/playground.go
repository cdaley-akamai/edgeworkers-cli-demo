package devserver

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
)

//go:embed playground.html
var playgroundHTML []byte

// invocationSchemaPlaceholder marks where the embedded invocation schema is inlined into
// playground.html, letting the page validate requests client-side without a schema fetch.
const invocationSchemaPlaceholder = "__INVOCATION_SCHEMA__"

var (
	playgroundPageOnce sync.Once
	playgroundPage     []byte
)

// renderedPlaygroundPage returns playground.html with invocationSchemaPlaceholder replaced
// by the embedded invocation schema JSON.
func renderedPlaygroundPage() []byte {
	playgroundPageOnce.Do(func() {
		playgroundPage = bytes.Replace(playgroundHTML, []byte(invocationSchemaPlaceholder), invocationSchemaFile, 1)
	})
	return playgroundPage
}

func newPlaygroundCmd(version string) *cobra.Command {
	var port int
	var noBrowser bool
	command := &cobra.Command{
		Use:   "playground",
		Short: "Start EdgeWorkersDevServer and serve the interactive playground page",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPlayground(cmd.Context(), version, port, !noBrowser, cmd.ErrOrStderr())
		},
	}
	command.Flags().IntVar(&port, "port", 8888, "port for the playground HTTP server")
	command.Flags().BoolVar(&noBrowser, "no-browser", false, "do not automatically open the playground page in a browser")
	return command
}

// runPlayground serves the playground page and the /run endpoint, blocking until ctx is
// canceled or the server fails. When openInBrowser is true, the playground page is opened
// in the OS default browser once the server is listening.
func runPlayground(ctx context.Context, version string, port int, openInBrowser bool, stderr io.Writer) error {
	logger := log.New(stderr, "", log.LstdFlags)
	mux := http.NewServeMux()
	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		handleRunRequest(w, r, logger)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(renderedPlaygroundPage())
	})
	server := &http.Server{Handler: mux}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", port, err)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- serveListenerOrNil(server, listener) }()

	logger.Printf("EdgeWorkersDevServer %s starting: playground port=%d", version, port)
	url := fmt.Sprintf("http://localhost:%d/", port)
	logger.Printf("EdgeWorkersDevServer %s ready: open %s", version, url)

	if openInBrowser {
		if err := openBrowser(url); err != nil {
			logger.Printf("could not open browser automatically: %v; open %s manually", err, url)
		}
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Printf("EdgeWorkersDevServer shutting down")
		_ = server.Close()
		return nil
	case err := <-errCh:
		return err
	}
}
