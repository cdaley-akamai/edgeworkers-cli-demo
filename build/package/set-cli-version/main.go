// Command set-cli-version updates one command's version field in cli.json in place.
//
// It patches only the "version" value belonging to the named command, leaving the
// rest of the file's formatting untouched, since cli.json's exact layout is a
// hand-maintained Akamai CLI plugin manifest.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

func main() {
	cliPath := flag.String("cli", "cli.json", "path to cli.json")
	command := flag.String("command", "", "command name to update, e.g. edgeworkers")
	version := flag.String("version", "", "new version value")
	flag.Parse()

	if *command == "" || *version == "" {
		fmt.Fprintln(os.Stderr, "command and version are required")
		os.Exit(1)
	}

	if err := run(*cliPath, *command, *version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cliPath, command, version string) error {
	data, err := os.ReadFile(cliPath)
	if err != nil {
		return fmt.Errorf("read cli.json: %w", err)
	}

	pattern := regexp.MustCompile(`("name":\s*"` + regexp.QuoteMeta(command) + `"[\s\S]*?"version":\s*")[^"]*(")`)
	if !pattern.Match(data) {
		return fmt.Errorf("command %q not found in %s", command, cliPath)
	}

	updated := pattern.ReplaceAll(data, []byte(`${1}`+version+`${2}`))
	return os.WriteFile(cliPath, updated, 0o644)
}
