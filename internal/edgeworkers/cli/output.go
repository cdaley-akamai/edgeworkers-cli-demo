package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
)

func decodeAPIResult(value, out any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode API result: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode API result: %w", err)
	}
	return nil
}

// renderJSON writes data as indented JSON to w.
func renderJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// outputData writes JSON to stdout when jsonOutput is enabled, otherwise it calls renderFn.
func outputData(data interface{}, jsonOutput bool, renderFn func()) error {
	if jsonOutput {
		return renderJSON(os.Stdout, data)
	}
	renderFn()
	return nil
}

// renderTable writes a tab-aligned table to w.
func renderTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	sep := make([]string, len(headers))
	for i, h := range headers {
		sep[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
}

// printAPIError writes a structured error to w.
func printAPIError(w io.Writer, operation string, err *edgeworkersapi.APIError, hint string) {
	fmt.Fprintf(w, "\nError: failed to %s\n", operation)
	fmt.Fprintf(w, "  HTTP status:  %d\n", err.StatusCode)
	if err.Title != "" {
		fmt.Fprintf(w, "  Title:        %s\n", err.Title)
	}
	if err.Detail != "" {
		fmt.Fprintf(w, "  API message:  %s\n", err.Detail)
	}
	if hint != "" {
		fmt.Fprintf(w, "  Hint:         %s\n", hint)
	}
	if err.Instance != "" {
		fmt.Fprintf(w, "  Request ID:   %s\n", err.Instance)
	}
	fmt.Fprintln(w, "\nRun with --debug to see full request and response details.")
}

// handleErr formats a structured command error for the executable to report.
func handleErr(operation string, err error, hint string) error {
	var output strings.Builder
	var apiErr *edgeworkersapi.APIError
	if ok := isAPIError(err, &apiErr); ok {
		printAPIError(&output, operation, apiErr, hint)
	} else {
		fmt.Fprintf(&output, "Error: failed to %s\n  %s", operation, err.Error())
		if hint != "" {
			fmt.Fprintf(&output, "\n  Hint: %s", hint)
		}
	}
	return fmt.Errorf("%s", strings.TrimSpace(output.String()))
}

func isAPIError(err error, target **edgeworkersapi.APIError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*edgeworkersapi.APIError); ok {
		*target = e
		return true
	}
	return false
}
