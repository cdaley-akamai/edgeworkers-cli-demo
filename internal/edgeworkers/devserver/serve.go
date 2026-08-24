package devserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func newServeCmd(version string) *cobra.Command {
	var port int
	var debugPort int
	command := &cobra.Command{
		Use:   "serve",
		Short: "Start EdgeWorkersDevServer as a long-lived HTTP server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd.Context(), version, port, debugPort, cmd.ErrOrStderr())
		},
	}
	command.Flags().IntVar(&port, "port", 9998, "port for the JSON request/response server")
	command.Flags().IntVar(&debugPort, "debug", 9999, "port for the debug protocol")
	return command
}

// runServe starts the JSON request server and the debug placeholder server, blocking until
// ctx is canceled or either server fails.
func runServe(ctx context.Context, version string, port, debugPort int, stderr io.Writer) error {
	logger := log.New(stderr, "", log.LstdFlags)
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: newRunMux(logger)}
	debugServer := &http.Server{Addr: fmt.Sprintf(":%d", debugPort), Handler: newDebugMux(logger)}

	errCh := make(chan error, 2)
	go func() { errCh <- serveOrNil(server) }()
	go func() { errCh <- serveOrNil(debugServer) }()

	logger.Printf("EdgeWorkersDevServer %s starting: port=%d debug=%d", version, port, debugPort)
	logger.Printf("EdgeWorkersDevServer %s ready for requests", version)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Printf("EdgeWorkersDevServer shutting down")
		_ = server.Close()
		_ = debugServer.Close()
		return nil
	case err := <-errCh:
		return err
	}
}

func serveOrNil(server *http.Server) error {
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// serveListenerOrNil serves server on an already-bound listener, treating a graceful
// shutdown (http.ErrServerClosed) as success.
func serveListenerOrNil(server *http.Server, listener net.Listener) error {
	err := server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func newRunMux(logger *log.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		handleRunRequest(w, r, logger)
	})
	return mux
}

func handleRunRequest(w http.ResponseWriter, r *http.Request, logger *log.Logger) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("error: read request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	response, err := handleRequest(body)
	if err != nil {
		logger.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		logger.Printf("error: write response: %v", err)
	}
}

// newDebugMux is a placeholder; the debug protocol is not yet defined.
func newDebugMux(logger *log.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("debug: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotImplemented)
	})
	return mux
}
