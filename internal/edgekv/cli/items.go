package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	edgekvapi "github.com/akamai/edgeworkers-cli/internal/edgekv/api"
	"github.com/spf13/cobra"
)

func resultMessage(result any) (string, bool) {
	if message, ok := result.(string); ok && message != "" {
		return message, true
	}
	if body, ok := result.(map[string]any); ok {
		message, ok := body["message"].(string)
		if ok && message != "" {
			return message, true
		}
	}
	return "", false
}

func newItemsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "items", Short: "Manage EdgeKV data items"}
	cmd.AddCommand(newItemsListCmd(), newItemsGetCmd(), newItemsPutCmd(), newItemsDeleteCmd())
	return cmd
}

func newItemsListCmd() *cobra.Command {
	var maxItemsValue int
	var sandboxID string
	cmd := &cobra.Command{
		Use:   "list <network> <namespace> <group-id>",
		Short: "List items in an EdgeKV data group",
		Args:  cobra.ExactArgs(3),
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
			var maxItems *float64
			if cmd.Flags().Changed("max-items") {
				value := float64(maxItemsValue)
				maxItems = &value
			}
			items, err := c.ListItems(cmd.Context(), args[0], args[1], args[2], maxItems, sandboxID)
			if err != nil {
				return apiError("list EdgeKV items", err)
			}
			return outputData(cmd, items, func(w io.Writer) {
				rows := make([][]string, 0, len(items))
				for _, item := range items {
					rows = append(rows, []string{item})
				}
				renderTable(w, []string{"ITEM ID"}, rows)
			})
		},
	}
	cmd.Flags().IntVar(&maxItemsValue, "max-items", 0, "maximum number of items to return")
	cmd.Flags().StringVar(&sandboxID, "sandbox-id", "", "sandbox ID for the data operation")
	return cmd
}

func newItemsGetCmd() *cobra.Command {
	var sandboxID string
	cmd := &cobra.Command{
		Use:   "get <network> <namespace> <group-id> <item-id>",
		Short: "Get an EdgeKV item",
		Args:  cobra.ExactArgs(4),
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
			data, err := c.GetItem(cmd.Context(), args[0], args[1], args[2], args[3], sandboxID, "")
			if err != nil {
				return apiError("get EdgeKV item", err)
			}
			item := string(data)
			return outputData(cmd, map[string]string{"value": item}, func(w io.Writer) { fmt.Fprintln(w, item) })
		},
	}
	cmd.Flags().StringVar(&sandboxID, "sandbox-id", "", "sandbox ID for the data operation")
	return cmd
}

func newItemsPutCmd() *cobra.Command {
	var sandboxID string
	cmd := &cobra.Command{
		Use:   "put <type> <network> <namespace> <group-id> <item-id> <value>",
		Short: "Create or update an EdgeKV item",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, contentType, err := itemValue(args[0], args[5])
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
			result, err := c.WriteItem(cmd.Context(), args[1], args[2], args[3], args[4], value, contentType, sandboxID)
			if err != nil {
				return apiError("put EdgeKV item", err)
			}
			if message, ok := resultMessage(result); ok {
				return outputData(cmd, map[string]string{"message": message}, func(w io.Writer) { fmt.Fprintln(w, message) })
			}
			return outputData(cmd, map[string]string{"item": args[4]}, func(w io.Writer) { fmt.Fprintf(w, "EdgeKV item %q updated.\n", args[4]) })
		},
	}
	cmd.Flags().StringVar(&sandboxID, "sandbox-id", "", "sandbox ID for the data operation")
	return cmd
}

func newItemsDeleteCmd() *cobra.Command {
	var sandboxID string
	cmd := &cobra.Command{
		Use:   "delete <network> <namespace> <group-id> <item-id>",
		Short: "Delete an EdgeKV item",
		Args:  cobra.ExactArgs(4),
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
			result, err := c.DeleteItem(cmd.Context(), args[0], args[1], args[2], args[3], sandboxID)
			if err != nil {
				return apiError("delete EdgeKV item", err)
			}
			if message, ok := resultMessage(result); ok {
				return outputData(cmd, map[string]string{"message": message}, func(w io.Writer) { fmt.Fprintln(w, message) })
			}
			return outputData(cmd, map[string]string{"item": args[3]}, func(w io.Writer) { fmt.Fprintf(w, "EdgeKV item %q deleted.\n", args[3]) })
		},
	}
	cmd.Flags().StringVar(&sandboxID, "sandbox-id", "", "sandbox ID for the data operation")
	return cmd
}

func itemValue(kind, value string) (string, string, error) {
	switch strings.ToLower(kind) {
	case "text":
		return value, "text/plain", nil
	case "jsonfile":
		data, err := os.ReadFile(value)
		if err != nil {
			return "", "", fmt.Errorf("read JSON file %q: %w", value, err)
		}
		if !jsonValid(data) {
			return "", "", fmt.Errorf("JSON file %q does not contain valid JSON", value)
		}
		return string(data), "application/json", nil
	default:
		return "", "", fmt.Errorf("type must be text or jsonfile")
	}
}

func jsonValid(data []byte) bool {
	var value interface{}
	return json.Unmarshal(data, &value) == nil
}
