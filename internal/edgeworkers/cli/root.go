// Package cli implements the Akamai EdgeWorkers command-line interface.
package cli

import (
	"time"

	"github.com/akamai/edgeworkers-cli/internal/cliconfig"
	edgeworkersmcp "github.com/akamai/edgeworkers-cli/internal/edgeworkers/mcp"
	"github.com/spf13/cobra"
)

var configStore = cliconfig.NewStore(cliconfig.DefaultPath())

const product = cliconfig.EdgeWorkers

// NewRootCmd creates an independent EdgeWorkers root command.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "edgeworkers",
		Short:         "Manage Akamai EdgeWorkers",
		Long:          "A CLI for managing Akamai EdgeWorkers. Works as both a Spin plugin and an Akamai CLI plugin.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.PersistentFlags().String("edgerc", "", "path to .edgerc file (default: ~/.edgerc)")
	cmd.PersistentFlags().String("edgerc-section", "default", ".edgerc section to use")
	cmd.PersistentFlags().String("cli-config-section", "default", "CLI configuration section for stored defaults")
	cmd.PersistentFlags().String("account-switch-key", "", "switch account context")
	cmd.PersistentFlags().Int("timeout", 120, "HTTP request timeout in seconds")
	cmd.PersistentFlags().Bool("debug", false, "print verbose request/response debug info")
	cmd.PersistentFlags().Bool("json", false, "output JSON to stdout")
	cmd.AddCommand(
		newIDsCmd(),
		newVersionsCmd(),
		newActivationsCmd(),
		newRevisionsCmd(),
		newAccountCmd(),
		newReportsCmd(),
		newDebugCmd(),
		newLoggingCmd(),
		cliconfig.NewConfigCmd(configStore),
		newDevCmd(),
		newMCPCmd(version),
	)
	return cmd
}

func jsonOutput(cmd *cobra.Command) bool {
	value, _ := cmd.Root().PersistentFlags().GetBool("json")
	return value
}

func newMCPCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run the Akamai EdgeWorkers MCP server over stdio",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return err
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			runtime, err := edgeworkersmcp.NewServer(edgeworkersmcp.Config{
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
