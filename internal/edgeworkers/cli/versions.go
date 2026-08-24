package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
)

//go:embed bundle.schema.json
var bundleSchemaFile []byte

var (
	bundleSchemaOnce sync.Once
	bundleSchema     *jsonschema.Schema
	bundleSchemaErr  error
)

// newVersionsCmd returns the versions command group.
func newVersionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "versions", Short: "Version and bundle management"}
	cmd.AddCommand(newVersionsListCmd(), newVersionsUploadCmd(), newVersionsDownloadCmd(), newVersionsDeleteCmd(), newVersionsValidateCmd())
	return cmd
}

func newVersionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list <ew-id>", Short: "List versions of an EdgeWorker ID", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
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
			raw, err := c.ListVersions(cmd.Context(), float64(ewID))
			if err != nil {

				return handleErr("list versions for EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.ListEdgeWorkerVersionsResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			rows := make([][]string, len(result.EdgeWorkerVersions))
			for i, v := range result.EdgeWorkerVersions {
				rows[i] = []string{v.Version, v.Checksum, v.CreatedBy, v.CreatedTime}
			}
			return outputData(result.EdgeWorkerVersions, jsonOutput(cmd), func() {
				renderTable(os.Stdout, []string{"VERSION", "CHECKSUM", "CREATED BY", "CREATED"}, rows)
			})
		},
	}
}

func newVersionsUploadCmd() *cobra.Command {
	var bundlePath string
	var directory string
	cmd := &cobra.Command{
		Use: "upload <ew-id>", Short: "Upload a new code bundle version", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if bundlePath == "" && directory == "" {
				return fmt.Errorf("one of --bundle or --directory is required")
			}
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
			}
			var data []byte
			if bundlePath != "" {
				data, err = os.ReadFile(bundlePath)
				if err != nil {
					return fmt.Errorf("read bundle %s: %w", bundlePath, err)
				}
			} else {
				data, err = bundleDirectory(directory)
				if err != nil {
					return err
				}
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
			raw, err := c.CreateVersion(cmd.Context(), float64(ewID), bytes.NewReader(data))
			if err != nil {

				return handleErr("upload version for EdgeWorker ID "+args[0], err, "")

			}
			result := &sdkew.EdgeWorkerVersion{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Version %s uploaded for EdgeWorker ID %d.\n", result.Version, result.EdgeWorkerID)
			})
		},
	}
	cmd.Flags().StringVar(&bundlePath, "bundle", "", "path to a packaged .tgz bundle")
	cmd.Flags().StringVar(&directory, "directory", "", "directory containing main.js and bundle.json to package")
	cmd.MarkFlagsMutuallyExclusive("bundle", "directory")
	return cmd
}

func bundleDirectory(directory string) ([]byte, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", directory, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", directory)
	}
	if err := validateBundleDirectory(directory); err != nil {
		return nil, err
	}

	var bundle bytes.Buffer
	gzipWriter := gzip.NewWriter(&bundle)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == directory || entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("bundle contains unsupported file %s", relativePath)
		}
		if info.Mode().Perm()&0111 != 0 {
			return fmt.Errorf("bundle contains executable file %s", relativePath)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relativePath)
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tarWriter, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}); err != nil {
		return nil, fmt.Errorf("package directory %s: %w", directory, err)
	}
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("close tarball: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("close compressed tarball: %w", err)
	}
	return bundle.Bytes(), nil
}

func validateBundleDirectory(directory string) error {
	mainPath := filepath.Join(directory, "main.js")
	mainInfo, err := os.Stat(mainPath)
	if err != nil {
		return fmt.Errorf("bundle must contain main.js: %w", err)
	}
	if !mainInfo.Mode().IsRegular() {
		return fmt.Errorf("main.js must be a regular file")
	}

	manifestPath := filepath.Join(directory, "bundle.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("bundle must contain bundle.json: %w", err)
	}
	return validateBundleManifest(manifestData)
}

func validateBundleManifest(manifestData []byte) error {
	schema, err := compiledBundleSchema()
	if err != nil {
		return err
	}
	var manifest any
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("parse bundle.json: %w", err)
	}
	if err := schema.Validate(manifest); err != nil {
		return fmt.Errorf("validate bundle.json: %w", err)
	}
	return nil
}

func compiledBundleSchema() (*jsonschema.Schema, error) {
	bundleSchemaOnce.Do(func() {
		var schemaDocument any
		if err := json.Unmarshal(bundleSchemaFile, &schemaDocument); err != nil {
			bundleSchemaErr = fmt.Errorf("parse embedded bundle schema: %w", err)
			return
		}
		compiler := jsonschema.NewCompiler()
		compiler.UseRegexpEngine(ecmaRegexpEngine)
		if err := compiler.AddResource("bundle.schema.json", schemaDocument); err != nil {
			bundleSchemaErr = fmt.Errorf("load embedded bundle schema: %w", err)
			return
		}
		bundleSchema, bundleSchemaErr = compiler.Compile("bundle.schema.json")
		if bundleSchemaErr != nil {
			bundleSchemaErr = fmt.Errorf("compile embedded bundle schema: %w", bundleSchemaErr)
		}
	})
	return bundleSchema, bundleSchemaErr
}

type ecmaRegexp regexp2.Regexp

func (r *ecmaRegexp) MatchString(value string) bool {
	matched, err := (*regexp2.Regexp)(r).MatchString(value)
	return err == nil && matched
}

func (r *ecmaRegexp) String() string {
	return (*regexp2.Regexp)(r).String()
}

func ecmaRegexpEngine(pattern string) (jsonschema.Regexp, error) {
	compiled, err := regexp2.Compile(pattern, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	return (*ecmaRegexp)(compiled), nil
}

func newVersionsDownloadCmd() *cobra.Command {
	var destPath string
	cmd := &cobra.Command{
		Use: "download <ew-id> <version-id>", Short: "Download a version's code bundle", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if destPath == "" {
				destPath = fmt.Sprintf("edgeworker-%s-v%s.tgz", args[0], args[1])
			}
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
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
			data, err := c.DownloadVersionContent(cmd.Context(), float64(ewID), args[1])
			if err != nil {

				return handleErr(fmt.Sprintf("download version %s for EdgeWorker ID %s", args[1], args[0]), err, "")

			}
			if err := os.WriteFile(destPath, data, 0o600); err != nil {
				return fmt.Errorf("write bundle to %s: %w", destPath, err)
			}
			fmt.Fprintf(os.Stdout, "Bundle saved to %s\n", destPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&destPath, "output", "", "destination file path")
	return cmd
}

func newVersionsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use: "delete <ew-id> <version-id>", Short: "Permanently delete a version", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ewID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("ew-id must be a number: %s", args[0])
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
			if err := c.DeleteVersion(cmd.Context(), float64(ewID), args[1]); err != nil {
				return handleErr(fmt.Sprintf("delete version %s for EdgeWorker ID %s", args[1], args[0]), err, "")
			}
			fmt.Fprintf(os.Stdout, "Version %s deleted.\n", args[1])
			return nil
		},
	}
}

func newVersionsValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "validate <bundle-path>", Short: "Validate a bundle without uploading", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read bundle %s: %w", args[0], err)
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
			raw, err := c.ValidateCodeBundle(cmd.Context(), bytes.NewReader(data))
			if err != nil {

				return handleErr("validate bundle "+args[0], err, "")

			}
			result := &sdkew.ValidateBundleResponse{}
			if err := decodeAPIResult(raw, result); err != nil {
				return err
			}
			return outputData(result, jsonOutput(cmd), func() {
				if len(result.Errors) == 0 && len(result.Warnings) == 0 {
					fmt.Fprintln(os.Stdout, "Bundle is valid.")
					return
				}
				for _, e := range result.Errors {
					fmt.Fprintf(os.Stdout, "ERROR: [%s] %s\n", e.Type, e.Message)
				}
				for _, w := range result.Warnings {
					fmt.Fprintf(os.Stdout, "WARN:  [%s] %s\n", w.Type, w.Message)
				}
			})
		},
	}
}
