package devserver

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Read a single JSON request from stdin and write the JSON response to stdout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			in := cmd.InOrStdin()
			if isInteractiveTerminal(in) {
				return fmt.Errorf("no input provided: pipe a JSON request to stdin, e.g. echo '{}' | %s run", cmd.Root().Name())
			}
			return runPipe(in, cmd.OutOrStdout())
		},
	}
}

// isInteractiveTerminal reports whether r is a terminal rather than a pipe, redirected
// file, or other non-interactive source, so "run" can fail fast instead of blocking
// forever waiting for input that will never arrive.
func isInteractiveTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// runPipe reads one JSON request from r and writes the JSON response to w.
func runPipe(r io.Reader, w io.Writer) error {
	input, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	response, err := handleRequest(input)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	return enc.Encode(response)
}
