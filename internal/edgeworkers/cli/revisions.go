package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/spf13/cobra"
)

type listRevisionsParams struct {
	VersionID       string
	ActivationID    string
	Network         string
	PinnedOnly      bool
	CurrentlyPinned bool
}

func parseRevisionEdgeWorkerID(value string) (float64, error) {
	edgeWorkerID, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("ew-id must be a number: %s", value)
	}
	return edgeWorkerID, nil
}

// newRevisionsCmd returns the revisions command group.
func newRevisionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "revisions", Short: "Immutable revision management"}
	cmd.AddCommand(
		newRevisionsListCmd(), newRevisionsGetCmd(), newRevisionsCompareCmd(),
		newRevisionsActivateCmd(), newRevisionsPinCmd(), newRevisionsUnpinCmd(),
		newRevisionsBOMCmd(), newRevisionsDownloadCmd(), newRevisionsActivationsCmd(),
	)
	return cmd
}

func newRevisionsListCmd() *cobra.Command {
	var p listRevisionsParams
	cmd := &cobra.Command{
		Use: "list <ew-id>", Short: "List revision history", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
			}
			options := edgeworkersapi.RevisionListOptions{}
			if p.VersionID != "" {
				options.Version = &p.VersionID
			}
			if p.ActivationID != "" {
				options.ActivationID = &p.ActivationID
			}
			if p.Network != "" {
				network := strings.ToUpper(p.Network)
				options.Network = &network
			}
			if p.PinnedOnly {
				options.PinnedOnly = &p.PinnedOnly
			}
			if p.CurrentlyPinned {
				options.CurrentlyPinned = &p.CurrentlyPinned
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
			raw, err := c.ListRevisions(cmd.Context(), edgeWorkerID, options)
			if err != nil {

				return handleErr("list revisions for EdgeWorker ID "+args[0], err, "")

			}
			var result edgeworkersapi.RevisionList
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			rows := make([][]string, len(result.Revisions))
			for i, r := range result.Revisions {
				rows[i] = []string{r.RevisionID, r.Version, r.Network, r.Status, r.CreatedTime}
			}
			return outputData(result.Revisions, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"REVISION ID", "VERSION", "NETWORK", "STATUS", "CREATED"}, rows)
			})
		},
	}
	cmd.Flags().StringVar(&p.VersionID, "version-id", "", "filter by version")
	cmd.Flags().StringVar(&p.ActivationID, "activation-id", "", "filter by activation ID")
	cmd.Flags().StringVar(&p.Network, "network", "", "filter by network")
	cmd.Flags().BoolVar(&p.PinnedOnly, "pinned-only", false, "show only pinned revisions")
	cmd.Flags().BoolVar(&p.CurrentlyPinned, "currently-pinned", false, "show currently pinned revisions")
	return cmd
}

func newRevisionsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use: "get <ew-id> <revision-id>", Short: "Get details for a revision", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
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
			raw, err := c.GetRevision(cmd.Context(), edgeWorkerID, args[1])
			if err != nil {

				return handleErr("get revision "+args[1], err, "")

			}
			var rev edgeworkersapi.Revision
			if err := decodeAPIResult(raw, &rev); err != nil {
				return err
			}
			return outputData(rev, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"REVISION ID", "VERSION", "NETWORK", "STATUS", "CREATED"},
					[][]string{{rev.RevisionID, rev.Version, rev.Network, rev.Status, rev.CreatedTime}})
			})
		},
	}
}

func newRevisionsCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use: "compare <ew-id> <rev1> <rev2>", Short: "Compare dependencies between two revisions", Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
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
			raw, err := c.CompareRevisions(cmd.Context(), edgeWorkerID, args[1], args[2])
			if err != nil {

				return handleErr(fmt.Sprintf("compare revisions %s and %s", args[1], args[2]), err, "")

			}
			var result edgeworkersapi.RevisionComparison
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Added: %d  Removed: %d  Changed: %d\n", len(result.Added), len(result.Removed), len(result.Changed))
			})
		},
	}
}

func newRevisionsActivateCmd() *cobra.Command {
	var note string
	cmd := &cobra.Command{
		Use: "activate <ew-id> <revision-id> <network>", Short: "Activate a revision on a network", Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
			}
			network := strings.ToUpper(args[2])
			request := edgeworkersapi.RevisionActivationRequest{RevisionID: args[1], Network: &network}
			if note != "" {
				request.Note = &note
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
			raw, err := c.ActivateRevision(cmd.Context(), edgeWorkerID, request)
			if err != nil {

				return handleErr("activate revision "+args[1], err, "")

			}
			var result edgeworkersapi.RevisionActivation
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Revision %s activation submitted on %s — status: %s\n", result.RevisionID, result.Network, result.Status)
			})
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "activation note")
	return cmd
}

func newRevisionsPinCmd() *cobra.Command {
	var note string
	cmd := &cobra.Command{
		Use: "pin <ew-id> <revision-id>", Short: "Pin an active revision", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
			}
			request := edgeworkersapi.RevisionPinRequest{}
			if note != "" {
				request.PinNote = &note
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
			raw, err := c.PinRevision(cmd.Context(), edgeWorkerID, args[1], request)
			if err != nil {

				return handleErr("pin revision "+args[1], err, "")

			}
			var result edgeworkersapi.PinResult
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Revision %s pinned: %v\n", result.RevisionID, result.Pinned)
			})
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "pin note")
	return cmd
}

func newRevisionsUnpinCmd() *cobra.Command {
	var note string
	cmd := &cobra.Command{
		Use: "unpin <ew-id> <revision-id>", Short: "Unpin a pinned revision", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
			}
			request := edgeworkersapi.RevisionUnpinRequest{}
			if note != "" {
				request.UnpinNote = &note
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
			raw, err := c.UnpinRevision(cmd.Context(), edgeWorkerID, args[1], request)
			if err != nil {

				return handleErr("unpin revision "+args[1], err, "")

			}
			var result edgeworkersapi.PinResult
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Revision %s unpinned.\n", result.RevisionID)
			})
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "unpin note")
	return cmd
}

func newRevisionsBOMCmd() *cobra.Command {
	var activeVersions, currentlyPinned bool
	cmd := &cobra.Command{
		Use: "bom <ew-id> <revision-id>", Short: "View composite bundle details (bill of materials)", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
			}
			options := edgeworkersapi.RevisionBOMOptions{}
			if activeVersions {
				options.IncludeActiveVersions = &activeVersions
			}
			if currentlyPinned {
				options.IncludeCurrentlyPinnedRevisions = &currentlyPinned
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
			raw, err := c.GetRevisionBOM(cmd.Context(), edgeWorkerID, args[1], options)
			if err != nil {

				return handleErr("get BOM for revision "+args[1], err, "")

			}
			var result edgeworkersapi.RevisionBOM
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Revision %s — %d packages\n", result.RevisionID, len(result.Packages))
			})
		},
	}
	cmd.Flags().BoolVar(&activeVersions, "active-versions", false, "include active versions")
	cmd.Flags().BoolVar(&currentlyPinned, "currently-pinned", false, "include currently pinned revisions")
	return cmd
}

func newRevisionsDownloadCmd() *cobra.Command {
	var destPath string
	cmd := &cobra.Command{
		Use: "download <ew-id> <revision-id>", Short: "Download the combined revision bundle", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
			}
			if destPath == "" {
				destPath = fmt.Sprintf("edgeworker-%s-rev-%s.tgz", args[0], args[1])
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
			data, err := c.DownloadRevisionContent(cmd.Context(), edgeWorkerID, args[1])
			if err != nil {

				return handleErr("download revision "+args[1], err, "")

			}
			if err := os.WriteFile(destPath, data, 0o600); err != nil {
				return fmt.Errorf("write revision bundle to %s: %w", destPath, err)
			}
			fmt.Fprintf(os.Stdout, "Revision bundle saved to %s\n", destPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&destPath, "output", "", "destination file path")
	return cmd
}

func newRevisionsActivationsCmd() *cobra.Command {
	return &cobra.Command{
		Use: "activations <ew-id>", Short: "List revision activation status", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			edgeWorkerID, err := parseRevisionEdgeWorkerID(args[0])
			if err != nil {
				return err
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
			raw, err := c.ListRevisionActivations(cmd.Context(), edgeWorkerID)
			if err != nil {

				return handleErr("list revision activations for EdgeWorker ID "+args[0], err, "")

			}
			var result edgeworkersapi.RevisionActivationList
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			rows := make([][]string, len(result.Activations))
			for i, a := range result.Activations {
				rows[i] = []string{a.RevisionID, a.Network, a.Status}
			}
			return outputData(result.Activations, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"REVISION ID", "NETWORK", "STATUS"}, rows)
			})
		},
	}
}
