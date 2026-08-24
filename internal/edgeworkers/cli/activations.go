package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/spf13/cobra"
)

// newActivationsCmd returns the activations command group.
func newActivationsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "activations", Short: "Deploy versions to networks"}
	cmd.AddCommand(newActivationsListCmd(), newActivationsActivateCmd(), newActivationsDeactivateCmd())
	return cmd
}

func newActivationsListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list <ew-id>", Short: "List activation status for an EdgeWorker ID", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
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
			raw, err := c.ListActivations(cmd.Context(), float64(ewID), nil, nil)
			if err != nil {

				return handleErr("list activations for EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.ListActivationsResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.Activations))
			for i, a := range result.Activations {
				rows[i] = []string{strconv.Itoa(a.ActivationID), a.Version, a.Network, a.Status, a.CreatedBy, a.CreatedTime}
			}
			return outputData(result.Activations, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"ACTIVATION ID", "VERSION", "NETWORK", "STATUS", "CREATED BY", "CREATED"}, rows)
			})
		},
	}
}

func newActivationsActivateCmd() *cobra.Command {
	var autoPin bool
	cmd := &cobra.Command{
		Use: "activate <ew-id> <network> <version-id>", Short: "Activate a version on a network (staging|production)", Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
			}
			request := edgeworkersapi.DeploymentRequest{
				Network: strings.ToUpper(args[1]),
				Version: args[2],
			}
			if cmd.Flags().Changed("auto-pin") {
				request.AutoPin = &autoPin
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
			raw, err := c.ActivateEdgeWorker(cmd.Context(), float64(ewID), request)
			if err != nil {

				return handleErr(fmt.Sprintf("activate version %s for EdgeWorker ID %s on %s", args[2], args[0], args[1]), err,
					"Verify your .edgerc credentials have the 'EdgeWorkers - View, Activate' capability.")

			}
			result := &sdkew.Activation{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Activation %d submitted: EdgeWorker %s v%s on %s — status: %s\n",
					result.ActivationID, args[0], result.Version, result.Network, result.Status)
			})
		},
	}
	cmd.Flags().BoolVar(&autoPin, "auto-pin", true, "pin initial revision automatically")
	return cmd
}

func newActivationsDeactivateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "deactivate <ew-id> <network> <version-id>", Short: "Deactivate a version on a network", Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
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
			raw, err := c.DeactivateEdgeWorker(cmd.Context(), float64(ewID), edgeworkersapi.DeploymentRequest{
				Network: strings.ToUpper(args[1]),
				Version: args[2],
			})
			if err != nil {

				return handleErr(fmt.Sprintf("deactivate version %s for EdgeWorker ID %s on %s", args[2], args[0], args[1]), err, "")

			}
			result := &sdkew.Deactivation{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Deactivation %d submitted: EdgeWorker %s v%s on %s — status: %s\n",
					result.DeactivationID, args[0], result.Version, result.Network, result.Status)
			})
		},
	}
}
