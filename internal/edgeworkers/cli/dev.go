package cli

import (
	"fmt"
	"os/exec"

	"github.com/akamai/edgeworkers-cli/internal/edgeworkers/devmanager"
	"github.com/spf13/cobra"
)

func newDevCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "dev",
		Short: "Manage local EdgeWorkers development tooling",
	}
	command.AddCommand(
		newDevRunCmd(),
		newDevServeCmd(),
		newDevPlaygroundCmd(),
		newDevUpdateCmd(),
		newDevInstallCmd(),
		newDevListCmd(),
	)
	return command
}

func newDevRunCmd() *cobra.Command {
	var serverVersion string
	command := &cobra.Command{
		Use:   "run",
		Short: "Run a request against the EdgeWorkers local development server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := resolveDevServerBinary(serverVersion)
			if err != nil {
				return handleErr("resolve EdgeWorkersDevServer", err, "run 'dev install <version>' or 'dev update'")
			}
			exe := exec.CommandContext(cmd.Context(), path, "run")
			exe.Stdin = cmd.InOrStdin()
			exe.Stdout = cmd.OutOrStdout()
			exe.Stderr = cmd.ErrOrStderr()
			if err := exe.Run(); err != nil {
				return handleErr("run EdgeWorkersDevServer", err, "")
			}
			return nil
		},
	}
	command.Flags().StringVar(&serverVersion, "server-version", "", "local development server version")
	return command
}

func newDevPlaygroundCmd() *cobra.Command {
	var serverVersion string
	var port int
	var noBrowser bool
	command := &cobra.Command{
		Use:   "playground",
		Short: "Launch a local dev server playground",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := resolveDevServerBinary(serverVersion)
			if err != nil {
				return handleErr("resolve EdgeWorkersDevServer", err, "run 'dev install <version>' or 'dev update'")
			}
			args := []string{"playground", "--port", fmt.Sprint(port)}
			if noBrowser {
				args = append(args, "--no-browser")
			}
			exe := exec.CommandContext(cmd.Context(), path, args...)
			exe.Stdout = cmd.OutOrStdout()
			exe.Stderr = cmd.ErrOrStderr()
			if err := exe.Run(); err != nil {
				return handleErr("run EdgeWorkersDevServer", err, "")
			}
			return nil
		},
	}
	command.Flags().StringVar(&serverVersion, "server-version", "", "local development server version")
	command.Flags().IntVar(&port, "port", 8888, "local development server playground http port")
	command.Flags().BoolVar(&noBrowser, "no-browser", false, "do not automatically open the playground page in a browser")
	return command
}

func newDevServeCmd() *cobra.Command {
	var serverVersion string
	var port int
	var debugPort int
	command := &cobra.Command{
		Use:   "serve",
		Short: "Start the EdgeWorkers local development environment as a server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := resolveDevServerBinary(serverVersion)
			if err != nil {
				return handleErr("resolve EdgeWorkersDevServer", err, "run 'dev install <version>' or 'dev update'")
			}
			exe := exec.CommandContext(cmd.Context(), path, "serve", "--port", fmt.Sprint(port), "--debug", fmt.Sprint(debugPort))
			exe.Stdout = cmd.OutOrStdout()
			exe.Stderr = cmd.ErrOrStderr()
			if err := exe.Run(); err != nil {
				return handleErr("run EdgeWorkersDevServer", err, "")
			}
			return nil
		},
	}
	command.Flags().StringVar(&serverVersion, "server-version", "", "local development server version")
	command.Flags().IntVar(&port, "port", 9998, "Port for the server.")
	command.Flags().IntVar(&debugPort, "debug", 9999, "Port for the debug protocol.")
	return command
}

// resolveDevServerBinary locates the installed EdgeWorkersDevServer binary for version
// (empty selects the locally recorded current version).
func resolveDevServerBinary(version string) (string, error) {
	manager, err := devmanager.NewManager()
	if err != nil {
		return "", err
	}
	return manager.ResolveBinary(version)
}
