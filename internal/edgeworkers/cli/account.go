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

// newAccountCmd returns the account command group.
func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "account", Short: "Account and org metadata (read-only)"}
	cmd.AddCommand(newAccountGroupsCmd(), newAccountContractsCmd(), newAccountPropertiesCmd(), newAccountResourceTiersCmd(), newAccountLimitsCmd())
	return cmd
}

func newAccountGroupsCmd() *cobra.Command {
	return &cobra.Command{
		Use: "groups [group-id]", Short: "List groups and permission capabilities", Args: cobra.MaximumNArgs(1),
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
			if len(args) == 1 {
				groupID, err := strconv.ParseFloat(args[0], 64)
				if err != nil {
					return fmt.Errorf("group-id must be a number: %s", args[0])
				}
				raw, err := c.GetGroup(cmd.Context(), groupID)
				if err != nil {

					return handleErr("get group "+args[0], err, "")

				}
				g := &sdkew.PermissionGroup{}
				if err := decodeAPIResult(raw, g); err != nil {
					return err
				}
				return outputData(g, jsonOutput(cmd), func() {
					renderTable(os.Stdout, []string{"GROUP ID", "NAME", "CAPABILITIES"},
						[][]string{{strconv.FormatInt(g.ID, 10), g.Name, strings.Join(g.Capabilities, ", ")}},
					)
				})
			}

			raw, err := c.ListGroups(cmd.Context())
			if err != nil {

				return handleErr("list groups", err, "")

			}
			result := &sdkew.ListPermissionGroupsResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.PermissionGroups))
			for i, g := range result.PermissionGroups {
				rows[i] = []string{strconv.FormatInt(g.ID, 10), g.Name, strings.Join(g.Capabilities, ", ")}
			}
			return outputData(result.PermissionGroups, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"GROUP ID", "NAME", "CAPABILITIES"}, rows)
			})
		},
	}
}

func newAccountContractsCmd() *cobra.Command {
	return &cobra.Command{
		Use: "contracts", Short: "List contracts for the account",
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
			raw, err := c.ListContracts(cmd.Context())
			if err != nil {

				return handleErr("list contracts", err, "")

			}
			result := &sdkew.ListContractsResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.ContractIDs))
			for i, id := range result.ContractIDs {
				rows[i] = []string{id}
			}
			return outputData(result.ContractIDs, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"CONTRACT ID"}, rows)
			})
		},
	}
}

func newAccountPropertiesCmd() *cobra.Command {
	var activeOnly bool
	cmd := &cobra.Command{
		Use: "properties <ew-id>", Short: "List properties associated with an EdgeWorker ID", Args: cobra.ExactArgs(1),
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
			raw, err := c.ListProperties(cmd.Context(), float64(ewID), activeOnly, false)
			if err != nil {

				return handleErr("list properties for EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.ListPropertiesResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.Properties))
			for i, p := range result.Properties {
				rows[i] = []string{
					strconv.FormatInt(p.ID, 10),
					p.Name,
					strconv.Itoa(p.StagingVersion),
					strconv.Itoa(p.ProductionVersion),
				}
			}
			return outputData(result.Properties, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"PROPERTY ID", "NAME", "STAGING VERSION", "PRODUCTION VERSION"}, rows)
			})
		},
	}
	cmd.Flags().BoolVar(&activeOnly, "active-only", false, "show only properties with active versions")
	return cmd
}

func newAccountResourceTiersCmd() *cobra.Command {
	return &cobra.Command{
		Use: "resource-tiers <contract-id>", Short: "List available resource tiers for a contract", Args: cobra.ExactArgs(1),
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
			raw, err := c.ListResourceTiers(cmd.Context(), &args[0])
			if err != nil {

				return handleErr("list resource tiers", err, "")

			}
			result := &sdkew.ListResourceTiersResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.ResourceTiers))
			for i, t := range result.ResourceTiers {
				rows[i] = []string{strconv.Itoa(t.ID), t.Name}
			}
			return outputData(result.ResourceTiers, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"TIER ID", "NAME"}, rows)
			})
		},
	}
}

// newAccountLimitsCmd shows EdgeWorker limits for a given EW ID via its resource tier.
func newAccountLimitsCmd() *cobra.Command {
	return &cobra.Command{
		Use: "limits <ew-id>", Short: "Show EdgeWorker limits for a given EdgeWorker ID", Args: cobra.ExactArgs(1),
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
			raw, err := c.GetEdgeWorkerResourceTier(cmd.Context(), float64(ewID))
			if err != nil {

				return handleErr("get limits for EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.ResourceTier{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.EdgeWorkerLimits))
			for i, l := range result.EdgeWorkerLimits {
				rows[i] = []string{l.LimitName, strconv.FormatInt(l.LimitValue, 10), l.LimitUnit}
			}
			return outputData(result.EdgeWorkerLimits, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"LIMIT", "VALUE", "UNIT"}, rows)
			})
		},
	}
}
