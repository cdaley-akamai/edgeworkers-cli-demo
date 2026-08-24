package cli

import (
	"fmt"
	"os"

	"github.com/akamai/edgeworkers-cli/internal/edgeworkers/devmanager"
	"github.com/akamai/edgeworkers-cli/internal/search"
	"github.com/spf13/cobra"
)

func newDevUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update the EdgeWorkers local development server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			manager, err := devmanager.NewManager()
			if err != nil {
				return handleErr("update EdgeWorkersDevServer", err, "")
			}
			manifest, err := manager.Manifest(cmd.Context())
			if err != nil {
				return handleErr("fetch EdgeWorkersDevServer manifest", err, "")
			}
			state, err := manager.State()
			if err != nil {
				return handleErr("update EdgeWorkersDevServer", err, "")
			}
			if state.Current == manifest.Current.Version {
				fmt.Fprintf(cmd.OutOrStdout(), "EdgeWorkersDevServer is already up to date (version %s)\n", state.Current)
				return nil
			}
			if err := manager.Install(cmd.Context(), manifest.Current); err != nil {
				return handleErr("install EdgeWorkersDevServer "+manifest.Current.Version, err, "")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated EdgeWorkersDevServer to version %s\n", manifest.Current.Version)
			return nil
		},
	}
}

func newDevInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <version>",
		Short: "Install an EdgeWorkers local development server version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			manager, err := devmanager.NewManager()
			if err != nil {
				return handleErr("install EdgeWorkersDevServer "+version, err, "")
			}
			manifest, err := manager.Manifest(cmd.Context())
			if err != nil {
				return handleErr("fetch EdgeWorkersDevServer manifest", err, "")
			}
			release, err := manifest.FindRelease(version)
			if err != nil {
				return handleErr("install EdgeWorkersDevServer "+version, err, "run 'dev list' to see available versions")
			}
			if err := manager.Install(cmd.Context(), release); err != nil {
				return handleErr("install EdgeWorkersDevServer "+version, err, "")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed EdgeWorkersDevServer %s\n", version)
			return nil
		},
	}
}

func newDevListCmd() *cobra.Command {
	var query string
	command := &cobra.Command{
		Use:   "list",
		Short: "List available EdgeWorkers local development server versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			manager, err := devmanager.NewManager()
			if err != nil {
				return handleErr("list EdgeWorkersDevServer versions", err, "")
			}
			manifest, err := manager.Manifest(cmd.Context())
			if err != nil {
				return handleErr("fetch EdgeWorkersDevServer manifest", err, "")
			}
			releases := append([]devmanager.Release{manifest.Current}, manifest.Previous...)
			releases = search.FuzzyFilter(releases, query, func(r devmanager.Release) []string {
				return []string{r.Version, r.ReleaseDate}
			})
			rows := make([][]string, len(releases))
			for i, release := range releases {
				current := ""
				if release.Version == manifest.Current.Version {
					current = "yes"
				}
				rows[i] = []string{release.Version, release.ReleaseDate, current}
			}
			return outputData(releases, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"VERSION", "RELEASE DATE", "CURRENT"}, rows)
			})
		},
	}
	command.Flags().StringVar(&query, "search", "", "filter versions with a fuzzy search")
	return command
}
