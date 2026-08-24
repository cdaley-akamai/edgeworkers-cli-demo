// Package cli implements the Akamai EdgeKV command-line interface.
package cli

import (
	"time"

	"github.com/akamai/edgeworkers-cli/internal/cliconfig"
	edgekvmcp "github.com/akamai/edgeworkers-cli/internal/edgekv/mcp"
	"github.com/spf13/cobra"
)

var configStore = cliconfig.NewStore(cliconfig.DefaultPath())

const product = cliconfig.EdgeKV

// NewRootCmd creates an independent EdgeKV root command.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "edgekv",
		Short:         "Manage Akamai EdgeKV",
		Long:          "A CLI for managing Akamai EdgeKV.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.PersistentFlags().String("edgerc", "", "path to .edgerc file (default: ~/.edgerc)")
	cmd.PersistentFlags().String("edgerc-section", "default", ".edgerc section to use")
	cmd.PersistentFlags().String("cli-config-section", "default", "CLI configuration section for stored defaults")
	cmd.PersistentFlags().String("account-switch-key", "", "switch account context")
	cmd.PersistentFlags().Int("timeout", 120, "HTTP request timeout in seconds")
	cmd.PersistentFlags().Bool("debug", false, "print verbose request and response debug info")
	cmd.PersistentFlags().Bool("json", false, "output JSON to stdout")
	cmd.AddCommand(
		newDatabaseCmd(),
		newNamespacesCmd(),
		newItemsCmd(),
		newTokensCmd(),
		cliconfig.NewConfigCmd(configStore),
		newMCPCmd(version),
	)
	return cmd
}

func newRootCmd() *cobra.Command {
	return NewRootCmd("dev")
}

func newMCPCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run the Akamai EdgeKV MCP server over stdio",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return err
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			runtime, err := edgekvmcp.NewServer(edgekvmcp.Config{
				Version:          version,
				Tools:            config.MCP.Tools,
				LogOutput:        cmd.ErrOrStderr(),
				EdgeRc:           config.EdgeRc,
				Section:          config.Section,
				AccountSwitchKey: config.AccountSwitchKey,
				Timeout:          time.Duration(config.TimeoutSeconds) * time.Second,
				Debug:            debug,
			})
			if err != nil {
				return err
			}
			return runtime.RunStdio(cmd.Context())
		},
	}
}
