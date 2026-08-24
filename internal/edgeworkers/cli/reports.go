package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/spf13/cobra"
)

// newReportsCmd returns the reports command group.
func newReportsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "reports", Short: "Usage and traffic reports"}
	cmd.AddCommand(newReportsListCmd(), newReportsGetCmd(), newReportsActiveCustomersCmd())
	return cmd
}

func newReportsListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List available report types",
		RunE: func(cmd *cobra.Command, args []string) error {
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			raw, err := c.ListReports(cmd.Context())
			if err != nil {

				return handleErr("list reports", err, "")

			}
			result := &sdkew.ListReportsResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.Reports))
			for i, r := range result.Reports {
				rows[i] = []string{strconv.Itoa(r.ReportID), r.Name, r.Description}
			}
			return outputData(result.Reports, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"ID", "NAME", "DESCRIPTION"}, rows)
			})
		},
	}
}

func newReportsGetCmd() *cobra.Command {
	var startDate, endDate, status, eventHandler string
	cmd := &cobra.Command{
		Use: "get <report-id> <ew-id>", Short: "Get an EdgeWorkers report", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("report-id must be a number: %s", args[0])
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			request := edgeworkersapi.ReportRequest{
				ReportID:   float64(reportID),
				Start:      startDate,
				End:        endDate,
				EdgeWorker: args[1],
				Status:     status,
			}
			if eventHandler != "" {
				request.EventHandler = &eventHandler
			}
			raw, err := c.GetReport(cmd.Context(), request)
			if err != nil {

				return handleErr(fmt.Sprintf("get report %d for EdgeWorker ID %s", reportID, args[1]), err, "")

			}

			// Report ID 1 has a distinct response shape.
			if reportID == 1 {
				result := &sdkew.GetSummaryReportResponse{}
				if err := decodeAPIResult(raw, result); err != nil {
					return err
				}
				return outputData(result, jsonOutput(cmd), func() {
					fmt.Fprintf(os.Stdout, "Report 1 (summary): %s to %s\n", result.Start, result.End)
				})
			}

			result := &sdkew.GetReportResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Report %d: %d records\n", result.ReportID, len(result.Data))
			})
		},
	}
	cmd.Flags().StringVar(&startDate, "start-date", "", "report start date (2006-01-02T15:04:05.999Z)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "report end date (2006-01-02T15:04:05.999Z)")
	cmd.Flags().StringVar(&status, "status", "", "filter by status")
	cmd.Flags().StringVar(&eventHandler, "event-handler", "", "filter by event handler")
	return cmd
}

func newReportsActiveCustomersCmd() *cobra.Command {
	return &cobra.Command{
		Use: "active-customers [ew-id]", Short: "List active customers for a partner EdgeWorker ID", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var edgeWorkerID *float64
			if len(args) > 0 {
				parsed, err := strconv.ParseFloat(args[0], 64)
				if err != nil {
					return fmt.Errorf("ew-id must be a number: %s", args[0])
				}
				edgeWorkerID = &parsed
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			result, err := c.ListActiveCustomers(cmd.Context(), edgeWorkerID)
			if err != nil {

				return handleErr("list active customers", err, "")

			}
			return outputData(result, jsonOutput(cmd), func() { fmt.Fprintln(os.Stdout, result) })
		},
	}
}
