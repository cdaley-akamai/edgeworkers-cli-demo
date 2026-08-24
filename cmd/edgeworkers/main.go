// Command edgeworkers is the Akamai CLI extension for managing EdgeWorkers.
package main

import (
	"fmt"
	"os"

	edgeworkerscli "github.com/akamai/edgeworkers-cli/internal/edgeworkers/cli"
)

var version = "dev"

func main() {
	if err := edgeworkerscli.NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
