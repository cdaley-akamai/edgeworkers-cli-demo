package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/akamai/edgeworkers-cli/internal/apiclient"
	"github.com/spf13/cobra"
)

func outputData(cmd *cobra.Command, data interface{}, render func(io.Writer)) error {
	jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
	render(cmd.OutOrStdout())
	return nil
}

func renderTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	separator := make([]string, len(headers))
	for i, header := range headers {
		separator[i] = strings.Repeat("-", len(header))
	}
	fmt.Fprintln(tw, strings.Join(separator, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	_ = tw.Flush()
}

func apiError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if apiErr, ok := err.(*apiclient.APIError); ok {
		if apiErr.Detail != "" {
			return fmt.Errorf("%s: HTTP %d: %s", operation, apiErr.StatusCode, apiErr.Detail)
		}
		return fmt.Errorf("%s: HTTP %d", operation, apiErr.StatusCode)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
