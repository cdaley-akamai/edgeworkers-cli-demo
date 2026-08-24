// Command edgekv is the Akamai CLI extension for managing EdgeKV.
package main

import (
	"fmt"
	"os"

	edgekvcli "github.com/akamai/edgeworkers-cli/internal/edgekv/cli"
)

var version = "dev"

func main() {
	if err := edgekvcli.NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
