package cliconfig

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestStoreLoadProductAcceptsJSON5(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	contents := `{
		// Profiles are direct top-level keys.
		work: {
			edgeworkers: {
				edgerc: '/credentials/edgeworkers.edgerc',
				section: 'edgeworkers',
				accountSwitchKey: 'A-CCT123',
				timeout: 45,
				mcp: { tools: ['list*', 'get*'], },
			},
			edgekv: {
				edgerc: '/credentials/edgekv.edgerc',
			},
		},
	}`
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))

	store := NewStore(path)
	config, err := store.LoadProduct("work", EdgeWorkers)
	require.NoError(t, err)
	require.Equal(t, ProductConfig{
		EdgeRc:           "/credentials/edgeworkers.edgerc",
		Section:          "edgeworkers",
		AccountSwitchKey: "A-CCT123",
		TimeoutSeconds:   45,
		MCP: MCPConfig{
			Tools: []string{"list*", "get*"},
		},
	}, config)
}

func TestStoreLoadProductUsesDefaultsWhenFileIsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	config, err := NewStore(path).LoadProduct("default", EdgeKV)
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	require.Equal(t, ProductConfig{
		EdgeRc:         filepath.Join(home, ".edgerc"),
		Section:        "default",
		TimeoutSeconds: 120,
		MCP: MCPConfig{
			Tools: []string{"*"},
		},
	}, config)
}

func TestStoreLoadEffectiveProductAppliesOnlyChangedFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		work: {
			edgekv: {
				edgerc: '/stored/.edgerc',
				section: 'stored',
				accountSwitchKey: 'stored-account',
				timeout: 45,
				mcp: {tools: ['get*']},
			},
		},
	}`), 0o600))

	command := &cobra.Command{Use: "edgekv"}
	command.PersistentFlags().String("edgerc", "", "")
	command.PersistentFlags().String("edgerc-section", "default", "")
	command.PersistentFlags().String("cli-config-section", "default", "")
	command.PersistentFlags().String("account-switch-key", "", "")
	command.PersistentFlags().Int("timeout", 120, "")
	require.NoError(t, command.PersistentFlags().Set("cli-config-section", "work"))
	require.NoError(t, command.PersistentFlags().Set("edgerc", "/flag/.edgerc"))
	require.NoError(t, command.PersistentFlags().Set("timeout", "90"))

	config, err := NewStore(path).LoadEffectiveProduct(command, EdgeKV)
	require.NoError(t, err)
	require.Equal(t, ProductConfig{
		EdgeRc:           "/flag/.edgerc",
		Section:          "stored",
		AccountSwitchKey: "stored-account",
		TimeoutSeconds:   90,
		MCP:              MCPConfig{Tools: []string{"get*"}},
	}, config)
}

func TestNewConfigCmdOpensExistingMalformedFileWithoutParsing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	malformed := []byte(`{ this is not valid JSON5`)
	require.NoError(t, os.WriteFile(path, malformed, 0o600))

	openedPath := filepath.Join(t.TempDir(), "opened-path")
	editor := editorTestCommand(t)
	t.Setenv(editorHelperArgumentsPathEnv, openedPath)
	t.Setenv("VISUAL", editor)
	t.Setenv("EDITOR", "")

	command := NewConfigCmd(NewStore(path))
	command.SetIn(strings.NewReader(""))
	command.SetOut(&bytes.Buffer{})
	command.SetErr(&bytes.Buffer{})
	require.NoError(t, command.Execute())

	opened, err := os.ReadFile(openedPath)
	require.NoError(t, err)
	require.Equal(t, path, string(opened))
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, malformed, written)
}

func TestNewConfigCmdIsLeaf(t *testing.T) {
	command := NewConfigCmd(NewStore(filepath.Join(t.TempDir(), "ew-config.json")))
	require.False(t, command.HasSubCommands())
	require.NotNil(t, command.RunE)
	command.SetArgs([]string{"extra"})
	require.Error(t, command.Execute())
}

func TestStoreEditCreatesCommentedTemplateWithPrivatePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".akamai-cli", "ew-config.json")
	editor := editorTestCommand(t)
	t.Setenv("VISUAL", editor)
	t.Setenv("EDITOR", "")

	require.NoError(t, NewStore(path).Edit(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}))

	if runtime.GOOS != "windows" {
		fileInfo, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
		directoryInfo, err := os.Stat(filepath.Dir(path))
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o700), directoryInfo.Mode().Perm())
	}

	written, err := os.ReadFile(path)
	require.NoError(t, err)
	text := string(written)
	require.Contains(t, text, "//")
	require.Contains(t, text, "default:")
	require.Contains(t, text, "edgeworkers:")
	require.Contains(t, text, "edgekv:")

	_, err = NewStore(path).LoadProduct("default", EdgeWorkers)
	require.NoError(t, err)
	_, err = NewStore(path).LoadProduct("default", EdgeKV)
	require.NoError(t, err)
}

func TestStoreEditPrefersVisualAndPreservesArgumentsAndStreams(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	require.NoError(t, os.WriteFile(path, []byte("{}\n"), 0o600))

	argumentsPath := filepath.Join(t.TempDir(), "arguments")
	visual := editorTestCommand(t, "--wait", "workspace config")
	t.Setenv(editorHelperArgumentsPathEnv, argumentsPath)
	t.Setenv(editorHelperEchoStdinEnv, "1")
	t.Setenv("VISUAL", visual)
	t.Setenv("EDITOR", filepath.Join(t.TempDir(), "missing-editor"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := NewStore(path).Edit(strings.NewReader("payload\n"), &stdout, &stderr)
	require.NoError(t, err)
	require.Equal(t, "stdout:payload", stdout.String())
	require.Equal(t, "stderr:payload", stderr.String())

	arguments, err := os.ReadFile(argumentsPath)
	require.NoError(t, err)
	require.Equal(t, []string{"--wait", "workspace config", path}, strings.Split(strings.TrimSpace(string(arguments)), "\n"))
}

func TestStoreEditUsesEditorWhenVisualIsUnset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	require.NoError(t, os.WriteFile(path, []byte("{}\n"), 0o600))

	openedPath := filepath.Join(t.TempDir(), "opened-path")
	editor := editorTestCommand(t)
	t.Setenv(editorHelperArgumentsPathEnv, openedPath)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editor)

	require.NoError(t, NewStore(path).Edit(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}))
	opened, err := os.ReadFile(openedPath)
	require.NoError(t, err)
	require.Equal(t, path, string(opened))
}

func TestStoreEditReturnsEditorErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	require.NoError(t, os.WriteFile(path, []byte("{}\n"), 0o600))

	t.Run("nonzero exit", func(t *testing.T) {
		editor := editorTestCommand(t)
		t.Setenv(editorHelperStderrEnv, "editor failed")
		t.Setenv(editorHelperFailEnv, "1")
		t.Setenv("VISUAL", editor)
		t.Setenv("EDITOR", "")

		var stderr bytes.Buffer
		err := NewStore(path).Edit(strings.NewReader(""), &bytes.Buffer{}, &stderr)
		require.ErrorContains(t, err, "exit status 23")
		require.Equal(t, "editor failed", stderr.String())
	})

	t.Run("launch", func(t *testing.T) {
		t.Setenv("VISUAL", filepath.Join(t.TempDir(), "missing-editor"))
		t.Setenv("EDITOR", "")

		err := NewStore(path).Edit(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
		require.ErrorContains(t, err, "editor")
	})
}

func TestDefaultEditor(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{goos: "darwin", want: "vi"},
		{goos: "linux", want: "vi"},
		{goos: "windows", want: "notepad"},
	}

	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			require.Equal(t, test.want, defaultEditor(test.goos))
		})
	}
}

func TestSplitEditorCommand(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    []string
		wantErr string
	}{
		{
			name:  "arguments",
			value: "code --wait --reuse-window",
			want:  []string{"code", "--wait", "--reuse-window"},
		},
		{
			name:  "quoted values",
			value: `code --profile "Work Config" 'literal value' ""`,
			want:  []string{"code", "--profile", "Work Config", "literal value", ""},
		},
		{
			name:    "malformed quotes",
			value:   `code "unterminated`,
			wantErr: "unterminated quote",
		},
		{
			name:    "empty command",
			value:   " \t ",
			wantErr: "empty command",
		},
		{
			name:  "backslashes and shell metacharacters",
			value: `"C:\Program Files\Editor\editor.exe" --literal '$HOME;$(touch hacked)&|<>*?' C:\config\file`,
			want:  []string{`C:\Program Files\Editor\editor.exe`, "--literal", `$HOME;$(touch hacked)&|<>*?`, `C:\config\file`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arguments, err := splitEditorCommand(test.value)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, arguments)
		})
	}
}

func TestStoreLoadProductRejectsNonPositiveTimeout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		default: {edgeworkers: {timeout: -1}},
	}`), 0o600))

	_, err := NewStore(path).LoadProduct("default", EdgeWorkers)
	require.ErrorContains(t, err, "timeout")
}

func TestStoreLoadProductPreservesExplicitEmptyToolList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ew-config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		default: {edgekv: {mcp: {tools: []}}},
	}`), 0o600))

	config, err := NewStore(path).LoadProduct("default", EdgeKV)
	require.NoError(t, err)
	require.NotNil(t, config.MCP.Tools)
	require.Empty(t, config.MCP.Tools)
	require.Equal(t, 120, config.TimeoutSeconds)
}

func TestDefaultPathUsesCanonicalJSONLocation(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".akamai-cli", "ew-config.json"), DefaultPath())
}

func TestDefaultPathReturnsEmptyWhenUserHomeIsUnavailable(t *testing.T) {
	clearUserHomeEnvironment(t)

	require.Empty(t, DefaultPath())
}

func TestStoreRejectsUnavailableConfigPath(t *testing.T) {
	store := NewStore("")

	t.Run("load product", func(t *testing.T) {
		_, err := store.LoadProduct("default", EdgeWorkers)
		require.ErrorContains(t, err, "user home/config path unavailable")
	})

	t.Run("edit", func(t *testing.T) {
		err := store.Edit(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
		require.ErrorContains(t, err, "user home/config path unavailable")
	})
}

func clearUserHomeEnvironment(t *testing.T) {
	t.Helper()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("USERPROFILE", "")
	case "plan9":
		t.Setenv("home", "")
	default:
		t.Setenv("HOME", "")
	}
}

const (
	editorHelperProcessEnv       = "CONFIG_EDITOR_TEST_PROCESS"
	editorHelperArgumentsPathEnv = "CONFIG_EDITOR_TEST_ARGUMENTS_PATH"
	editorHelperEchoStdinEnv     = "CONFIG_EDITOR_TEST_ECHO_STDIN"
	editorHelperStderrEnv        = "CONFIG_EDITOR_TEST_STDERR"
	editorHelperFailEnv          = "CONFIG_EDITOR_TEST_FAIL"
)

func TestEditorProcess(t *testing.T) {
	if os.Getenv(editorHelperProcessEnv) != "1" {
		return
	}

	arguments := editorProcessArguments(os.Args)
	if path := os.Getenv(editorHelperArgumentsPathEnv); path != "" {
		if err := os.WriteFile(path, []byte(strings.Join(arguments, "\n")), 0o600); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "write editor arguments: %v", err)
			os.Exit(2)
		}
	}
	if os.Getenv(editorHelperEchoStdinEnv) == "1" {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "read editor stdin: %v", err)
			os.Exit(2)
		}
		value := strings.TrimSuffix(string(input), "\n")
		_, _ = fmt.Fprintf(os.Stdout, "stdout:%s", value)
		_, _ = fmt.Fprintf(os.Stderr, "stderr:%s", value)
	}
	if value := os.Getenv(editorHelperStderrEnv); value != "" {
		_, _ = io.WriteString(os.Stderr, value)
	}
	if os.Getenv(editorHelperFailEnv) == "1" {
		os.Exit(23)
	}
	os.Exit(0)
}

func editorTestCommand(t *testing.T, arguments ...string) string {
	t.Helper()
	t.Setenv(editorHelperProcessEnv, "1")
	command := []string{
		quoteEditorArgument(os.Args[0]),
		"-test.run=^TestEditorProcess$",
		"--",
	}
	for _, argument := range arguments {
		command = append(command, quoteEditorArgument(argument))
	}
	return strings.Join(command, " ")
}

func quoteEditorArgument(argument string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(argument)
	return `"` + escaped + `"`
}

func editorProcessArguments(arguments []string) []string {
	for index, argument := range arguments {
		if argument == "--" {
			return arguments[index+1:]
		}
	}
	return nil
}
