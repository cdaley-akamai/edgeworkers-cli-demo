package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/search"
	"github.com/spf13/cobra"
)

// newIDsCmd returns the ids command group.
func newIDsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ids",
		Short: "EdgeWorker ID management",
	}
	cmd.AddCommand(
		newIDsListCmd(),
		newIDsCreateCmd(),
		newIDsUpdateCmd(),
		newIDsDeleteCmd(),
		newIDsCloneCmd(),
		newIDsResourceTierCmd(),
	)
	return cmd
}

func newIDsListCmd() *cobra.Command {
	var groupID, resourceTierID int
	var searchQuery string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered EdgeWorker IDs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			var groupFilter, resourceTierFilter *float64
			if groupID != 0 {
				value := float64(groupID)
				groupFilter = &value
			}
			if resourceTierID != 0 {
				value := float64(resourceTierID)
				resourceTierFilter = &value
			}
			raw, err := c.ListEdgeWorkers(cmd.Context(), groupFilter, resourceTierFilter)
			if err != nil {

				return handleErr("list EdgeWorker IDs", err, "")

			}
			result := &sdkew.ListEdgeWorkersIDResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			ids := search.FuzzyFilter(result.EdgeWorkers, searchQuery, func(id sdkew.EdgeWorkerID) []string {
				return []string{strconv.Itoa(id.EdgeWorkerID), id.Name}
			})
			rows := make([][]string, len(ids))
			for i, id := range ids {
				rows[i] = []string{strconv.Itoa(id.EdgeWorkerID), id.Name, strconv.FormatInt(id.GroupID, 10), strconv.Itoa(id.ResourceTierID)}
			}
			return outputData(ids, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"EW ID", "NAME", "GROUP ID", "RESOURCE TIER"}, rows)
			})
		},
	}
	cmd.Flags().IntVar(&groupID, "group-id", 0, "filter by group ID")
	cmd.Flags().IntVar(&resourceTierID, "resource-tier-id", 0, "filter by resource tier ID")
	cmd.Flags().StringVar(&searchQuery, "search", "", "fuzzy search by EdgeWorker ID or name")
	return cmd
}

func newIDsCreateCmd() *cobra.Command {
	var resourceTierID int
	cmd := &cobra.Command{
		Use:   "create <group-id> <name>",
		Short: "Register a new EdgeWorker ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("group-id must be a number: %s", args[0])
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			resourceTier := float64(resourceTierID)
			raw, err := c.CreateEdgeWorker(cmd.Context(), edgeworkersapi.EdgeWorkerMutation{
				Name:           args[1],
				GroupID:        float64(groupID),
				ResourceTierID: &resourceTier,
			})
			if err != nil {

				return handleErr(fmt.Sprintf("create EdgeWorker ID %q in group %s", args[1], args[0]), err, "")

			}
			result := &sdkew.EdgeWorkerID{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "EdgeWorker ID %d created: %s\n", result.EdgeWorkerID, result.Name)
			})
		},
	}
	cmd.Flags().IntVar(&resourceTierID, "resource-tier-id", 0, "resource tier to assign")
	return cmd
}

func newIDsUpdateCmd() *cobra.Command {
	var resourceTierID int
	cmd := &cobra.Command{
		Use:   "update <ew-id> <group-id> <name>",
		Short: "Update an EdgeWorker ID's group or name",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
			}
			groupID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("group-id must be a number: %s", args[1])
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			resourceTier := float64(resourceTierID)
			raw, err := c.UpdateEdgeWorker(cmd.Context(), float64(ewID), edgeworkersapi.EdgeWorkerMutation{
				Name:           args[2],
				GroupID:        float64(groupID),
				ResourceTierID: &resourceTier,
			})
			if err != nil {

				return handleErr("update EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.EdgeWorkerID{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "EdgeWorker ID %d updated.\n", result.EdgeWorkerID)
			})
		},
	}
	cmd.Flags().IntVar(&resourceTierID, "resource-tier-id", 0, "new resource tier ID")
	return cmd
}

func newIDsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <ew-id>",
		Short: "Permanently delete an EdgeWorker ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			if err := c.DeleteEdgeWorker(cmd.Context(), float64(ewID)); err != nil {
				return handleErr("delete EdgeWorker ID "+args[0], err, "")
			}
			fmt.Fprintf(os.Stdout, "EdgeWorker ID %s deleted.\n", args[0])
			return nil
		},
	}
}

func newIDsCloneCmd() *cobra.Command {
	var groupID int
	var name string
	cmd := &cobra.Command{
		Use:   "clone <ew-id> <resource-tier-id>",
		Short: "Clone an EdgeWorker ID to a new resource tier",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
			}
			rtID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("resource-tier-id must be a number: %s", args[1])
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			resourceTier := float64(rtID)
			raw, err := c.CloneEdgeWorker(cmd.Context(), float64(ewID), edgeworkersapi.EdgeWorkerMutation{
				ResourceTierID: &resourceTier,
				GroupID:        float64(groupID),
				Name:           name,
			})
			if err != nil {

				return handleErr("clone EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.EdgeWorkerID{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "EdgeWorker ID %s cloned to new ID %d.\n", args[0], result.EdgeWorkerID)
			})
		},
	}
	cmd.Flags().IntVar(&groupID, "group-id", 0, "group for the cloned ID")
	cmd.Flags().StringVar(&name, "name", "", "name for the cloned ID")
	return cmd
}

func newIDsResourceTierCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resource-tier <ew-id>",
		Short: "Show resource tier and limits for an EdgeWorker ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "Check --edgerc and --section flags.")
			}
			raw, err := c.GetEdgeWorkerResourceTier(cmd.Context(), float64(ewID))
			if err != nil {

				return handleErr("get resource tier for EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.ResourceTier{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				renderTable(os.Stdout,
					[]string{"RESOURCE TIER ID", "NAME"},
					[][]string{{strconv.Itoa(result.ID), result.Name}},
				)
				if len(result.EdgeWorkerLimits) > 0 {
					fmt.Fprintln(os.Stdout)
					limitRows := make([][]string, len(result.EdgeWorkerLimits))
					for i, l := range result.EdgeWorkerLimits {
						limitRows[i] = []string{l.LimitName, strconv.FormatInt(l.LimitValue, 10), l.LimitUnit}
					}
					renderTable(os.Stdout, []string{"LIMIT", "VALUE", "UNIT"}, limitRows)
				}
			})
		},
	}
}
