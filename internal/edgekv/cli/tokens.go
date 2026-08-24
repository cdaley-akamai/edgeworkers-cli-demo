package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	edgekvapi "github.com/akamai/edgeworkers-cli/internal/edgekv/api"
	"github.com/spf13/cobra"
)

func newTokensCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "tokens", Short: "Manage EdgeKV access tokens"}
	cmd.AddCommand(newTokensListCmd(), newTokensCreateCmd(), newTokensDownloadCmd(), newTokensRefreshCmd(), newTokensRevokeCmd())
	return cmd
}

func newTokensListCmd() *cobra.Command {
	var includeExpired bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List EdgeKV access tokens",
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
			var includeExpiredValue *bool
			if includeExpired {
				includeExpiredValue = &includeExpired
			}
			result, err := c.ListTokens(cmd.Context(), includeExpiredValue)
			if err != nil {
				return apiError("list EdgeKV tokens", err)
			}
			return outputData(cmd, result.Tokens, func(w io.Writer) {
				rows := make([][]string, 0, len(result.Tokens))
				for _, token := range result.Tokens {
					rows = append(rows, []string{
						stringMapValue(token, "name"),
						stringMapValue(token, "uuid"),
						stringMapValue(token, "expiry"),
						stringMapValue(token, "tokenActivationStatus"),
					})
				}
				renderTable(w, []string{"NAME", "UUID", "EXPIRY", "STATUS"}, rows)
			})
		},
	}
	cmd.Flags().BoolVar(&includeExpired, "include-expired", false, "include expired tokens")
	return cmd
}

func newTokensCreateCmd() *cobra.Command {
	var staging, production, edgeworkerIDs, namespaces, expiry, savePath string
	var overwrite bool
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an EdgeKV access token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			permissions, err := parseTokenPermissions(namespaces)
			if err != nil {
				return err
			}
			allowStaging, err := accessAllowed(staging)
			if err != nil {
				return fmt.Errorf("staging: %w", err)
			}
			allowProduction, err := accessAllowed(production)
			if err != nil {
				return fmt.Errorf("production: %w", err)
			}
			if !allowStaging && !allowProduction {
				return fmt.Errorf("at least one of staging or production must be allow")
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
			request := edgekvapi.CreateTokenRequest{
				Name:                 args[0],
				AllowOnStaging:       allowStaging,
				AllowOnProduction:    allowProduction,
				NamespacePermissions: permissionCodes(permissions),
			}
			if ids := edgeworkerIDList(edgeworkerIDs); ids != nil {
				request.RestrictToEdgeWorkerIDs = ids
			}
			if expiry != "" {
				parsedExpiry, err := time.Parse("2006-01-02", expiry)
				if err != nil {
					return fmt.Errorf("expiry must use YYYY-MM-DD")
				}
				request.Expiry = parsedExpiry.UTC().Format(time.RFC3339)
			}
			result, err := c.CreateToken(cmd.Context(), request)
			if err != nil {
				return apiError("create EdgeKV token", err)
			}
			return saveOrOutputToken(cmd, result, args[0], "created", savePath, overwrite)
		},
	}
	cmd.Flags().StringVar(&staging, "staging", "", "allow or deny token access on staging")
	cmd.Flags().StringVar(&production, "production", "", "allow or deny token access on production")
	cmd.Flags().StringVar(&edgeworkerIDs, "edgeworker-ids", "all", "comma-separated EdgeWorker IDs or all")
	cmd.Flags().StringVar(&namespaces, "namespaces", "", "namespace+permissions pairs, such as default+rwd,marketing+r")
	cmd.Flags().StringVar(&expiry, "expiry", "", "token expiry date in ISO 8601 format")
	cmd.Flags().StringVar(&savePath, "save-path", "", "directory, edgekv_tokens.js file, or .tgz bundle for token storage")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "replace an existing saved token for the same namespace")
	_ = cmd.MarkFlagRequired("staging")
	_ = cmd.MarkFlagRequired("production")
	_ = cmd.MarkFlagRequired("namespaces")
	return cmd
}

func newTokensDownloadCmd() *cobra.Command {
	var output, savePath string
	var overwrite bool
	cmd := &cobra.Command{
		Use:   "download <name>",
		Short: "Download an EdgeKV access token",
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
			result, err := c.GetToken(cmd.Context(), args[0])
			if err != nil {
				return apiError("download EdgeKV token", err)
			}
			if output != "" && savePath != "" {
				return fmt.Errorf("only one of output or save-path may be set")
			}
			if output != "" {
				return writeToken(cmd, result, output)
			}
			return saveOrOutputToken(cmd, result, args[0], "downloaded", savePath, overwrite)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "write token response to this file with mode 0600")
	cmd.Flags().StringVar(&savePath, "save-path", "", "directory, edgekv_tokens.js file, or .tgz bundle for token storage")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "replace an existing saved token for the same namespace")
	return cmd
}

func newTokensRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh <name>",
		Short: "Refresh an EdgeKV access token",
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
			result, err := c.RefreshToken(cmd.Context(), args[0])
			if err != nil {
				return apiError("refresh EdgeKV token", err)
			}
			return tokenOutput(cmd, result, args[0], "refreshed")
		},
	}
}

func newTokensRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <name>",
		Short: "Revoke an EdgeKV access token",
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
			result, err := c.DeleteToken(cmd.Context(), args[0])
			if err != nil {
				return apiError("revoke EdgeKV token", err)
			}
			name := stringMapValue(result, "name")
			if name == "" {
				name = args[0]
			}
			return outputData(cmd, result, func(w io.Writer) { fmt.Fprintf(w, "EdgeKV token %q revoked.\n", name) })
		},
	}
}

func permissionCodes(permissions sdkew.NamespacePermissions) map[string][]string {
	result := make(map[string][]string, len(permissions))
	for namespace, namespacePermissions := range permissions {
		codes := make([]string, 0, len(namespacePermissions))
		for _, permission := range namespacePermissions {
			switch permission {
			case sdkew.PermissionRead:
				codes = append(codes, "r")
			case sdkew.PermissionWrite:
				codes = append(codes, "w")
			case sdkew.PermissionDelete:
				codes = append(codes, "d")
			}
		}
		result[namespace] = codes
	}
	return result
}

func stringMapValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprintf("%v", value)
}

func parseTokenPermissions(value string) (sdkew.NamespacePermissions, error) {
	permissions := sdkew.NamespacePermissions{}
	for _, pair := range strings.Split(value, ",") {
		name, rawPermissions, found := strings.Cut(strings.TrimSpace(pair), "+")
		if !found || name == "" || rawPermissions == "" {
			return nil, fmt.Errorf("namespaces must use namespace+permissions pairs")
		}
		set := make([]sdkew.Permission, 0, len(rawPermissions))
		seen := map[rune]bool{}
		for _, permission := range rawPermissions {
			if seen[permission] {
				continue
			}
			seen[permission] = true
			switch permission {
			case 'r':
				set = append(set, sdkew.PermissionRead)
			case 'w':
				set = append(set, sdkew.PermissionWrite)
			case 'd':
				set = append(set, sdkew.PermissionDelete)
			default:
				return nil, fmt.Errorf("invalid permission %q for namespace %q; use r, w, or d", string(permission), name)
			}
		}
		permissions[name] = set
	}
	return permissions, nil
}

func accessAllowed(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "allow":
		return true, nil
	case "deny":
		return false, nil
	default:
		return false, fmt.Errorf("must be allow or deny")
	}
}

func edgeworkerIDList(value string) []string {
	if strings.EqualFold(value, "all") || value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func tokenOutput(cmd *cobra.Command, result map[string]interface{}, name, action string) error {
	jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
	if jsonOutput {
		return outputData(cmd, result, func(io.Writer) {})
	}
	if value, ok := result["value"].(string); ok && value != "" {
		fmt.Fprintln(cmd.OutOrStdout(), value)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "EdgeKV token %q %s.\n", name, action)
	return nil
}

func saveOrOutputToken(cmd *cobra.Command, result map[string]interface{}, name, action, savePath string, overwrite bool) error {
	if savePath == "" {
		return tokenOutput(cmd, result, name, action)
	}
	if err := saveToken(savePath, result, overwrite); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "EdgeKV token %q saved to %s.\n", name, savePath)
	return nil
}

func writeToken(cmd *cobra.Command, result map[string]interface{}, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode EdgeKV token: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write token file %q: %w", path, err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "EdgeKV token written to %s.\n", path)
	return nil
}

type savedToken struct {
	Name      string `json:"name"`
	Value     string `json:"value,omitempty"`
	Reference string `json:"reference,omitempty"`
}

func saveToken(path string, result map[string]interface{}, overwrite bool) error {
	namespaces, err := tokenNamespaces(result)
	if err != nil {
		return err
	}
	token, err := savedTokenValue(result)
	if err != nil {
		return err
	}
	if strings.EqualFold(filepath.Ext(path), ".tgz") {
		return saveTokenBundle(path, namespaces, token, overwrite)
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, "edgekv_tokens.js")
	} else if os.IsNotExist(err) && filepath.Ext(path) == "" {
		return fmt.Errorf("token directory %q does not exist", path)
	}
	return saveTokenFile(path, namespaces, token, overwrite)
}

func tokenNamespaces(result map[string]interface{}) ([]string, error) {
	namespaces := make([]string, 0)
	if values, ok := result["namespacePermissions"].(map[string]interface{}); ok {
		for name := range values {
			namespaces = append(namespaces, name)
		}
	}
	if value, ok := result["value"].(string); ok && value != "" {
		parts := strings.Split(value, ".")
		if len(parts) == 3 {
			payload, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err == nil {
				var claims map[string]interface{}
				if json.Unmarshal(payload, &claims) == nil {
					for key := range claims {
						if strings.HasPrefix(key, "namespace-") {
							namespaces = append(namespaces, key)
						}
					}
				}
			}
		}
	}
	namespaces = uniqueStrings(namespaces)
	if len(namespaces) == 0 {
		return nil, fmt.Errorf("token response does not contain namespace permissions")
	}
	return namespaces, nil
}

func savedTokenValue(result map[string]interface{}) (savedToken, error) {
	name, _ := result["name"].(string)
	if name == "" {
		return savedToken{}, fmt.Errorf("token response does not contain a name")
	}
	if value, ok := result["value"].(string); ok && value != "" {
		return savedToken{Name: name, Value: value}, nil
	}
	if reference, ok := result["reference"].(string); ok && reference != "" {
		return savedToken{Name: name, Reference: reference}, nil
	}
	if uuid, ok := result["uuid"].(string); ok && uuid != "" {
		return savedToken{Name: name, Reference: uuid}, nil
	}
	return savedToken{}, fmt.Errorf("token response does not contain a value or reference")
}

func saveTokenFile(path string, namespaces []string, token savedToken, overwrite bool) error {
	content := map[string]savedToken{}
	if data, err := os.ReadFile(path); err == nil {
		content, err = parseTokenFile(data)
		if err != nil {
			return fmt.Errorf("read token file %q: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read token file %q: %w", path, err)
	}
	if err := mergeToken(content, namespaces, token, overwrite); err != nil {
		return err
	}
	data, err := formatTokenFile(content)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func saveTokenBundle(path string, namespaces []string, token savedToken, overwrite bool) error {
	input, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open token bundle %q: %w", path, err)
	}
	defer input.Close()

	gzipReader, err := gzip.NewReader(input)
	if err != nil {
		return fmt.Errorf("read token bundle %q: %w", path, err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	temporary, err := os.CreateTemp(filepath.Dir(path), ".edgekv-*.tgz")
	if err != nil {
		return fmt.Errorf("create token bundle: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	gzipWriter := gzip.NewWriter(temporary)
	tarWriter := tar.NewWriter(gzipWriter)
	found := false
	for {
		header, readErr := tarReader.Next()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read token bundle %q: %w", path, readErr)
		}
		if header.Name == "edgekv_tokens.js" {
			found = true
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return fmt.Errorf("read edgekv_tokens.js: %w", err)
			}
			content, err := parseTokenFile(data)
			if err != nil {
				return fmt.Errorf("read edgekv_tokens.js: %w", err)
			}
			if err := mergeToken(content, namespaces, token, overwrite); err != nil {
				return err
			}
			data, err = formatTokenFile(content)
			if err != nil {
				return err
			}
			header.Size = int64(len(data))
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}
			if _, err := tarWriter.Write(data); err != nil {
				return err
			}
			continue
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := io.Copy(tarWriter, tarReader); err != nil {
			return err
		}
	}
	if !found {
		content := map[string]savedToken{}
		if err := mergeToken(content, namespaces, token, overwrite); err != nil {
			return err
		}
		data, err := formatTokenFile(content)
		if err != nil {
			return err
		}
		if err := tarWriter.WriteHeader(&tar.Header{Name: "edgekv_tokens.js", Mode: 0600, Size: int64(len(data))}); err != nil {
			return err
		}
		if _, err := tarWriter.Write(data); err != nil {
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func mergeToken(content map[string]savedToken, namespaces []string, token savedToken, overwrite bool) error {
	for _, namespace := range namespaces {
		if existing, ok := content[namespace]; ok && !overwrite {
			if existing != token {
				return fmt.Errorf("namespace %q already has a different token; use --overwrite to replace it", namespace)
			}
			return fmt.Errorf("namespace %q already has this token; use --overwrite to replace it", namespace)
		}
		content[namespace] = token
	}
	return nil
}

func parseTokenFile(data []byte) (map[string]savedToken, error) {
	const prefix = "var edgekv_access_tokens = "
	start := bytes.Index(data, []byte(prefix))
	if start < 0 {
		return nil, fmt.Errorf("expected edgekv_access_tokens object")
	}
	start += len(prefix)
	end := bytes.IndexByte(data[start:], ';')
	if end < 0 {
		return nil, fmt.Errorf("expected edgekv_access_tokens assignment")
	}
	content := map[string]savedToken{}
	if err := json.Unmarshal(data[start:start+end], &content); err != nil {
		return nil, fmt.Errorf("parse token object: %w", err)
	}
	return content, nil
}

func formatTokenFile(content map[string]savedToken) ([]byte, error) {
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode token file: %w", err)
	}
	return []byte("var edgekv_access_tokens = " + string(data) + ";\nexport { edgekv_access_tokens };\n"), nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
