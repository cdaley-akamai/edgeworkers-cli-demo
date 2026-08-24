// Package cliconfig loads persistent settings shared by the EdgeWorkers and EdgeKV CLIs.
package cliconfig

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"github.com/alchemy/json5"
	"github.com/spf13/cobra"
)

// Product identifies a product subtree in a CLI configuration profile.
type Product string

const (
	// EdgeWorkers identifies the EdgeWorkers configuration subtree.
	EdgeWorkers Product = "edgeworkers"
	// EdgeKV identifies the EdgeKV configuration subtree.
	EdgeKV Product = "edgekv"
)

// MCPConfig contains product-specific MCP server settings.
type MCPConfig struct {
	// Tools contains the tool-name glob patterns enabled for the product server.
	Tools []string `json:"tools"`
}

// ProductConfig contains persistent defaults for one product in one profile.
type ProductConfig struct {
	// EdgeRc is the path to the .edgerc credentials file.
	EdgeRc string `json:"edgerc"`
	// Section selects the credentials section in EdgeRc.
	Section string `json:"section"`
	// AccountSwitchKey identifies the Akamai account for API requests.
	AccountSwitchKey string `json:"accountSwitchKey"`
	// TimeoutSeconds is the HTTP request timeout in seconds.
	TimeoutSeconds int `json:"timeout"`
	// MCP contains settings for the product MCP server.
	MCP MCPConfig `json:"mcp"`
}

type storedMCPConfig struct {
	Tools *[]string `json:"tools"`
}

type storedProductConfig struct {
	EdgeRc           *string          `json:"edgerc"`
	Section          *string          `json:"section"`
	AccountSwitchKey *string          `json:"accountSwitchKey"`
	TimeoutSeconds   *int             `json:"timeout"`
	MCP              *storedMCPConfig `json:"mcp"`
}

type storedProfile struct {
	EdgeWorkers *storedProductConfig `json:"edgeworkers"`
	EdgeKV      *storedProductConfig `json:"edgekv"`
}

// Store accesses CLI configuration in a JSON5 document.
type Store struct {
	path string
}

const configTemplate = `{
	// Each top-level key is a profile shared by both CLIs.
	default: {
		edgeworkers: {
			// edgerc: "/path/to/.edgerc",
			// section: "default",
			// accountSwitchKey: "",
			// timeout: 120,
			// mcp: { tools: ["*"] },
		},
		edgekv: {
			// edgerc: "/path/to/.edgerc",
			// section: "default",
			// accountSwitchKey: "",
			// timeout: 120,
			// mcp: { tools: ["*"] },
		},
	},
}
`

// DefaultPath returns the canonical shared CLI configuration path, or an empty
// string when the user's home directory is unavailable.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".akamai-cli", "ew-config.json")
}

// NewStore returns a configuration store backed by path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// LoadProduct loads one product's settings from a named top-level profile.
func (s *Store) LoadProduct(profileName string, product Product) (ProductConfig, error) {
	if err := s.validatePath(); err != nil {
		return ProductConfig{}, err
	}
	if err := validateProduct(product); err != nil {
		return ProductConfig{}, err
	}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return defaultProductConfig(), nil
	}
	if err != nil {
		return ProductConfig{}, fmt.Errorf("read config: %w", err)
	}

	var profiles map[string]storedProfile
	if err := json5.Unmarshal(data, &profiles); err != nil {
		return ProductConfig{}, fmt.Errorf("parse config: %w", err)
	}
	profile, ok := profiles[profileName]
	if !ok {
		return defaultProductConfig(), nil
	}
	var stored *storedProductConfig
	if product == EdgeWorkers {
		stored = profile.EdgeWorkers
	} else {
		stored = profile.EdgeKV
	}
	if stored == nil {
		return defaultProductConfig(), nil
	}
	return resolveProductConfig(*stored)
}

// LoadEffectiveProduct loads the selected product profile and applies root
// persistent flags that were explicitly set for the command invocation.
func (s *Store) LoadEffectiveProduct(command *cobra.Command, product Product) (ProductConfig, error) {
	flags := command.Root().PersistentFlags()
	profileName := "default"
	if flags.Lookup("cli-config-section") != nil {
		profileName, _ = flags.GetString("cli-config-section")
	}
	config, err := s.LoadProduct(profileName, product)
	if err != nil {
		return ProductConfig{}, err
	}
	if flag := flags.Lookup("edgerc"); flag != nil && flag.Changed {
		config.EdgeRc, _ = flags.GetString("edgerc")
	}
	if flag := flags.Lookup("edgerc-section"); flag != nil && flag.Changed {
		config.Section, _ = flags.GetString("edgerc-section")
	}
	if flag := flags.Lookup("account-switch-key"); flag != nil && flag.Changed {
		config.AccountSwitchKey, _ = flags.GetString("account-switch-key")
	}
	if flag := flags.Lookup("timeout"); flag != nil && flag.Changed {
		config.TimeoutSeconds, _ = flags.GetInt("timeout")
	}
	return config, nil
}

// NewConfigCmd creates the shared config editor command for store.
func NewConfigCmd(store *Store) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Edit persistent CLI defaults",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return store.Edit(command.InOrStdin(), command.OutOrStdout(), command.ErrOrStderr())
		},
	}
}

// Edit creates a commented configuration template when the file is missing,
// then opens the file in the configured editor and waits for it to exit. Edit
// passes the supplied streams to the editor and does not parse the file.
func (s *Store) Edit(stdin io.Reader, stdout, stderr io.Writer) error {
	if err := s.validatePath(); err != nil {
		return err
	}
	if err := s.ensureFile(); err != nil {
		return err
	}

	editor, err := editorCommand()
	if err != nil {
		return err
	}
	command := exec.Command(editor[0], append(editor[1:], s.path)...)
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("run editor %q: %w", editor[0], err)
	}
	return nil
}

func (s *Store) validatePath() error {
	if s.path == "" {
		return fmt.Errorf("user home/config path unavailable")
	}
	return nil
}

func (s *Store) ensureFile() error {
	if _, err := os.Stat(s.path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat config: %w", err)
	}

	directory := filepath.Dir(s.path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	file, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create config: %w", err)
	}
	removeIncomplete := true
	defer func() {
		if removeIncomplete {
			_ = os.Remove(s.path)
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("set config permissions: %w", err)
	}
	if _, err := io.WriteString(file, configTemplate); err != nil {
		_ = file.Close()
		return fmt.Errorf("write config template: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close config: %w", err)
	}
	removeIncomplete = false
	return nil
}

func editorCommand() ([]string, error) {
	value := strings.TrimSpace(os.Getenv("VISUAL"))
	if value == "" {
		value = strings.TrimSpace(os.Getenv("EDITOR"))
	}
	if value == "" {
		value = defaultEditor(runtime.GOOS)
	}
	arguments, err := splitEditorCommand(value)
	if err != nil {
		return nil, fmt.Errorf("parse editor command: %w", err)
	}
	return arguments, nil
}

func defaultEditor(goos string) string {
	if goos == "windows" {
		return "notepad"
	}
	return "vi"
}

func splitEditorCommand(value string) ([]string, error) {
	runes := []rune(value)
	arguments := make([]string, 0, 2)
	var current strings.Builder
	var quote rune
	started := false

	for index := 0; index < len(runes); index++ {
		character := runes[index]
		switch {
		case character == '\\' && quote != '\'':
			if index+1 < len(runes) {
				next := runes[index+1]
				if unicode.IsSpace(next) || next == '\\' || next == '\'' || next == '"' {
					current.WriteRune(next)
					started = true
					index++
					continue
				}
			}
			current.WriteRune(character)
			started = true
		case character == '\'' || character == '"':
			switch quote {
			case 0:
				quote = character
				started = true
			case character:
				quote = 0
			default:
				current.WriteRune(character)
			}
		case unicode.IsSpace(character) && quote == 0:
			if started {
				arguments = append(arguments, current.String())
				current.Reset()
				started = false
			}
		default:
			current.WriteRune(character)
			started = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote")
	}
	if started {
		arguments = append(arguments, current.String())
	}
	if len(arguments) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return arguments, nil
}

func validateProduct(product Product) error {
	if product != EdgeWorkers && product != EdgeKV {
		return fmt.Errorf("unknown product %q", product)
	}
	return nil
}

func defaultProductConfig() ProductConfig {
	edgerc := ".edgerc"
	if home, err := os.UserHomeDir(); err == nil {
		edgerc = filepath.Join(home, ".edgerc")
	}
	return ProductConfig{
		EdgeRc:         edgerc,
		Section:        "default",
		TimeoutSeconds: 120,
		MCP: MCPConfig{
			Tools: []string{"*"},
		},
	}
}

func resolveProductConfig(stored storedProductConfig) (ProductConfig, error) {
	config := defaultProductConfig()
	if stored.EdgeRc != nil && *stored.EdgeRc != "" {
		config.EdgeRc = *stored.EdgeRc
	}
	if stored.Section != nil && *stored.Section != "" {
		config.Section = *stored.Section
	}
	if stored.AccountSwitchKey != nil {
		config.AccountSwitchKey = *stored.AccountSwitchKey
	}
	if stored.TimeoutSeconds != nil {
		if *stored.TimeoutSeconds <= 0 {
			return ProductConfig{}, fmt.Errorf("timeout must be a positive integer number of seconds")
		}
		config.TimeoutSeconds = *stored.TimeoutSeconds
	}
	if stored.MCP != nil && stored.MCP.Tools != nil {
		config.MCP.Tools = make([]string, len(*stored.MCP.Tools))
		copy(config.MCP.Tools, *stored.MCP.Tools)
	}
	return config, nil
}
