package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/spf13/cobra"
)

// newLoggingCmd returns the logging command group.
func newLoggingCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "logging", Short: "Manage EdgeWorkers logging overrides"}
	cmd.AddCommand(newLoggingGetCmd(), newLoggingSetCmd())
	return cmd
}

func newLoggingGetCmd() *cobra.Command {
	return &cobra.Command{
		Use: "get <ew-id> [logging-id]", Short: "Get an EdgeWorkers logging override", Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
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
			var raw any
			if len(args) > 1 {
				loggingID, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return fmt.Errorf("logging-id must be a number: %s", args[1])
				}
				raw, err = c.GetLoggingOverride(cmd.Context(), edgeWorkerID, loggingID)
			} else {
				raw, err = c.ListLoggingOverrides(cmd.Context(), edgeWorkerID)
			}
			if err != nil {

				return handleErr("get log level for EdgeWorker ID "+args[0], err, "")

			}
			var result edgeworkersapi.LogLevelList
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			rows := make([][]string, len(result.Loggings))
			for i, l := range result.Loggings {
				rows[i] = []string{l.LoggingID, l.Level, l.Network, l.ExpiresAt}
			}
			return outputData(result.Loggings, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"LOGGING ID", "LEVEL", "NETWORK", "EXPIRES"}, rows)
			})
		},
	}
}

func newLoggingSetCmd() *cobra.Command {
	return &cobra.Command{
		Use: "set <ew-id> <network> <level>", Short: "Set an EdgeWorkers logging override", Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
			}
			request := edgeworkersapi.LoggingOverrideRequest{
				Level:   strings.ToUpper(args[2]),
				Network: strings.ToUpper(args[1]),
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
			raw, err := c.CreateLoggingOverride(cmd.Context(), edgeWorkerID, request)
			if err != nil {

				return handleErr(fmt.Sprintf("set log level %s for EdgeWorker ID %s on %s", args[2], args[0], args[1]), err, "")

			}
			var result edgeworkersapi.LogLevel
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Log level set to %s on %s (ID: %s)\n", result.Level, result.Network, result.LoggingID)
			})
		},
	}
}
