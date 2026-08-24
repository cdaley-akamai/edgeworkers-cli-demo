package cli

import (
	"fmt"
	"io"
	"strconv"
	"time"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	edgekvapi "github.com/akamai/edgeworkers-cli/internal/edgekv/api"
	"github.com/spf13/cobra"
)

func newNamespacesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "namespaces", Short: "Manage EdgeKV namespaces"}
	cmd.AddCommand(
		newNamespacesListCmd(), newNamespacesGetCmd(), newNamespacesCreateCmd(), newNamespacesUpdateCmd(),
		newNamespaceGroupsCmd(), newNamespacePermissionsCmd(),
	)
	return cmd
}

func newNamespacesListCmd() *cobra.Command {
	var details bool
	cmd := &cobra.Command{
		Use:   "list <network>",
		Short: "List EdgeKV namespaces",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgekvapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			result, err := c.ListNamespaces(cmd.Context(), args[0], details)
			if err != nil {
				return apiError("list EdgeKV namespaces", err)
			}
			return outputData(cmd, result, func(w io.Writer) {
				rows := make([][]string, 0, len(result))
				for _, namespace := range result {
					rows = append(rows, namespaceRow(namespace))
				}
				renderTable(w, []string{"NAMESPACE", "RETENTION", "GROUP ID", "GEO LOCATION"}, rows)
			})
		},
	}
	cmd.Flags().BoolVar(&details, "details", false, "include namespace attributes")
	return cmd
}

func newNamespacesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <network> <namespace>",
		Short: "Get an EdgeKV namespace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgekvapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			result, err := c.GetNamespace(cmd.Context(), args[0], args[1])
			if err != nil {
				return apiError("get EdgeKV namespace", err)
			}
			return outputData(cmd, result, func(w io.Writer) {
				renderTable(w, []string{"NAMESPACE", "STATUS", "RETENTION", "GROUP ID", "GEO LOCATION"}, [][]string{{result.Name, result.NamespaceStatus, seconds(result.Retention), integer(result.GroupID), result.GeoLocation}})
			})
		},
	}
}

func newNamespacesCreateCmd() *cobra.Command {
	var retentionDays, groupID int
	var geoLocation, policy string
	cmd := &cobra.Command{
		Use:   "create <network> <namespace>",
		Short: "Create an EdgeKV namespace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			retention, err := retentionSeconds(retentionDays)
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("group-id") {
				return fmt.Errorf("group-id is required; use 0 to allow all account groups")
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgekvapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			body := map[string]any{"namespace": args[1], "retentionInSeconds": retention, "groupId": groupID}
			if geoLocation != "" {
				body["geoLocation"] = geoLocation
			}
			if policy != "" {
				parsed, parseErr := parseNamespacePolicy(policy)
				if parseErr != nil {
					return parseErr
				}
				body["dataAccessPolicy"] = parsed
			}
			result, err := c.CreateNamespace(cmd.Context(), args[0], body)
			if err != nil {
				return apiError("create EdgeKV namespace", err)
			}
			return outputData(cmd, result, func(w io.Writer) { fmt.Fprintf(w, "EdgeKV namespace %q created.\n", result.Name) })
		},
	}
	cmd.Flags().IntVar(&retentionDays, "retention-days", 0, "retention period in days; 0 keeps data indefinitely")
	cmd.Flags().IntVar(&groupID, "group-id", 0, "Akamai access group ID")
	cmd.Flags().StringVar(&geoLocation, "geo-location", "", "persistent storage location")
	cmd.Flags().StringVar(&policy, "data-access-policy", "", "restrictDataAccess=BOOL")
	_ = cmd.MarkFlagRequired("retention-days")
	return cmd
}

func newNamespacesUpdateCmd() *cobra.Command {
	var retentionDays int
	var geoLocation string
	cmd := &cobra.Command{
		Use:   "update <network> <namespace>",
		Short: "Update an EdgeKV namespace retention period",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			retention, err := retentionSeconds(retentionDays)
			if err != nil {
				return err
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgekvapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			if args[1] == "default" {
				return fmt.Errorf("cannot update retention for the default namespace")
			}
			current, err := c.GetNamespace(cmd.Context(), args[0], args[1])
			if err != nil {
				return apiError("get EdgeKV namespace", err)
			}
			body := map[string]any{"namespace": args[1], "retentionInSeconds": retention}
			if current.GroupID != nil {
				body["groupId"] = *current.GroupID
			}
			if geoLocation != "" {
				body["geoLocation"] = geoLocation
			} else if current.GeoLocation != "" {
				body["geoLocation"] = current.GeoLocation
			}
			result, err := c.UpdateNamespace(cmd.Context(), args[0], args[1], body)
			if err != nil {
				return apiError("update EdgeKV namespace", err)
			}
			return outputData(cmd, result, func(w io.Writer) { fmt.Fprintf(w, "EdgeKV namespace %q updated.\n", args[1]) })
		},
	}
	cmd.Flags().IntVar(&retentionDays, "retention-days", 0, "retention period in days; 0 keeps data indefinitely")
	cmd.Flags().StringVar(&geoLocation, "geo-location", "", "persistent storage location")
	_ = cmd.MarkFlagRequired("retention-days")
	return cmd
}

func newNamespaceGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "groups", Short: "List data groups in an EdgeKV namespace"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list <network> <namespace>",
		Short: "List data groups in a namespace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgekvapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return apiError("connect to Akamai API", err)
			}
			result, err := c.ListNamespaceGroups(cmd.Context(), args[0], args[1])
			if err != nil {
				return apiError("list EdgeKV data groups", err)
			}
			return outputData(cmd, result, func(w io.Writer) {
				rows := make([][]string, 0, len(result))
				for _, group := range result {
					rows = append(rows, []string{group})
				}
				renderTable(w, []string{"GROUP ID"}, rows)
			})
		},
	})
	return cmd
}

func newNamespacePermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "permissions", Short: "Manage namespace access groups"}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "set <namespace> <group-id>",
			Short: "Set an EdgeKV namespace access group",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				groupID, err := strconv.Atoi(args[1])
				if err != nil || groupID < 0 {
					return fmt.Errorf("group-id must be a non-negative number")
				}
				productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
				if err != nil {
					return apiError("connect to Akamai API", err)
				}
				debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
				c, err := edgekvapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
				if err != nil {
					return apiError("connect to Akamai API", err)
				}
				result, err := c.ReauthorizeNamespace(cmd.Context(), args[0], float64(groupID))
				if err != nil {
					return apiError("set EdgeKV namespace access group", err)
				}
				return outputData(cmd, result, func(w io.Writer) { fmt.Fprintf(w, "EdgeKV namespace %q access group updated.\n", args[0]) })
			},
		},
	)
	return cmd
}

func retentionSeconds(days int) (int, error) {
	if days < 0 || days > 3650 {
		return 0, fmt.Errorf("retention-days must be between 0 and 3650")
	}
	return days * 24 * 60 * 60, nil
}

func namespaceRow(namespace sdkew.Namespace) []string {
	return []string{namespace.Name, seconds(namespace.Retention), integer(namespace.GroupID), namespace.GeoLocation}
}

func seconds(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func integer(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}
