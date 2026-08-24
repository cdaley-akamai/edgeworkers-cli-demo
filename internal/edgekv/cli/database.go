package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	edgekvapi "github.com/akamai/edgeworkers-cli/internal/edgekv/api"
	"github.com/spf13/cobra"
)

func newDatabaseCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "database", Short: "Manage the EdgeKV database"}
	cmd.AddCommand(newDatabaseInitializeCmd(), newDatabaseStatusCmd(), newDatabaseSetPolicyCmd())
	return cmd
}

func newDatabaseInitializeCmd() *cobra.Command {
	var policy string
	cmd := &cobra.Command{
		Use:   "initialize",
		Short: "Initialize the EdgeKV database",
		Args:  cobra.NoArgs,
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
			var parsedPolicy *edgekvapi.DataAccessPolicy
			if policy != "" {
				parsed, parseErr := parseDatabasePolicy(policy)
				if parseErr != nil {
					return parseErr
				}
				parsedPolicy = &parsed
			}
			result, err := c.InitializeDatabase(cmd.Context(), parsedPolicy)
			if err != nil {
				return apiError("initialize EdgeKV database", err)
			}
			return outputData(cmd, result, func(w io.Writer) {
				fmt.Fprintln(w, "EdgeKV database initialization started.")
			})
		},
	}
	cmd.Flags().StringVar(&policy, "data-access-policy", "", "restrictDataAccess=BOOL,allowNamespacePolicyOverride=BOOL")
	return cmd
}

func newDatabaseStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show EdgeKV database initialization status",
		Args:  cobra.NoArgs,
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
			result, err := c.GetDatabaseStatus(cmd.Context())
			if err != nil {
				return apiError("get EdgeKV database status", err)
			}
			return outputData(cmd, result, func(w io.Writer) {
				renderTable(w, []string{"ACCOUNT", "STAGING", "PRODUCTION", "CP CODE"}, [][]string{{result.AccountStatus, result.StagingStatus, result.ProductionStatus, result.CPCode}})
			})
		},
	}
}

func newDatabaseSetPolicyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-policy <data-access-policy>",
		Short: "Set the default database data access policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := parseDatabasePolicy(args[0])
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
			result, err := c.UpdateDatabasePolicy(cmd.Context(), parsed)
			if err != nil {
				return apiError("set EdgeKV database policy", err)
			}
			return outputData(cmd, result, func(w io.Writer) {
				fmt.Fprintln(w, "EdgeKV database data access policy updated.")
			})
		},
	}
}

func parseDatabasePolicy(value string) (edgekvapi.DataAccessPolicy, error) {
	values, err := parseBooleanPairs(value, "restrictDataAccess", "allowNamespacePolicyOverride")
	if err != nil {
		return edgekvapi.DataAccessPolicy{}, err
	}
	return edgekvapi.DataAccessPolicy{
		RestrictDataAccess:           values["restrictDataAccess"],
		AllowNamespacePolicyOverride: values["allowNamespacePolicyOverride"],
	}, nil
}

func parseNamespacePolicy(value string) (map[string]bool, error) {
	return parseBooleanPairs(value, "restrictDataAccess")
}

func parseBooleanPairs(value string, required ...string) (map[string]bool, error) {
	result := make(map[string]bool, len(required))
	allowed := make(map[string]bool, len(required))
	for _, key := range required {
		allowed[key] = true
	}
	for _, pair := range strings.Split(value, ",") {
		key, raw, found := strings.Cut(pair, "=")
		if !found {
			return nil, fmt.Errorf("data access policy must use key=value pairs")
		}
		key = strings.TrimSpace(key)
		if !allowed[key] {
			return nil, fmt.Errorf("data access policy does not support %q", key)
		}
		boolValue, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("data access policy %q must be true or false", key)
		}
		result[key] = boolValue
	}
	for _, key := range required {
		if _, ok := result[key]; !ok {
			return nil, fmt.Errorf("data access policy requires %q", key)
		}
	}
	return result, nil
}
