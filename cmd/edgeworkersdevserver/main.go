// Command edgeworkersdevserver is the EdgeWorkersDevServer local development runtime.
package main

import (
	"fmt"
	"os"

	"github.com/akamai/edgeworkers-cli/internal/edgeworkers/devserver"
)

var version = "dev"

func main() {
	if err := devserver.NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
