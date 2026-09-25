// Command mycli is the entry point. It holds no logic: it calls cli.Execute and
// exits with the code the command tree returns.
package main

import (
	"os"

	"github.com/example/mycli/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
