// Package devserver implements the EdgeWorkersDevServer executable's own CLI.
//
// EdgeWorkersDevServer is a separate binary that the edgeworkers CLI's "dev"
// commands exec as a subprocess to run, serve, or preview EdgeWorkers locally.
package devserver

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the EdgeWorkersDevServer root command.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "edgeworkersdevserver",
		Short:         fmt.Sprintf("EdgeWorkersDevServer %s: local EdgeWorkers development runtime", version),
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.AddCommand(
		newRunCmd(),
		newServeCmd(version),
		newPlaygroundCmd(version),
	)
	return cmd
}
