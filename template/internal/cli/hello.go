package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// maxNameLen bounds the greeting input so the command has a real validation
// failure to demonstrate.
const maxNameLen = 64

// newHelloCmd is the worked example of the command pattern. Copy it when you
// add a command, then delete it. It shows all five pieces:
//
//  1. run(g, "<name>", ...) wraps the handler in the shared output contract
//  2. the handler returns a *cmdResult, never writes to stdout itself
//  3. failures return a typed *cliError, which fixes the exit code
//  4. cmdResult.data feeds --json; cmdResult.plain feeds human stdout
//  5. warnings ride along in the envelope without failing the command
func newHelloCmd(g *globalFlags) *cobra.Command {
	var shout bool

	cmd := &cobra.Command{
		Use:   "hello [name]",
		Short: "Greet someone",
		Long: "Greet someone by name.\n\n" +
			"READS    the name argument, else $MYCLI_NAME\n" +
			"WRITES   nothing\n" +
			"NEVER    makes a network call",
		Args: cobra.MaximumNArgs(1),
		RunE: run(g, "hello", func(_ *cobra.Command, args []string) (*cmdResult, error) {
			name := os.Getenv("MYCLI_NAME")
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				return nil, prereqError(
					"no name given: pass one as an argument or set MYCLI_NAME", nil)
			}
			if len(name) > maxNameLen {
				return nil, validationError(
					fmt.Sprintf("name is %d bytes; the limit is %d", len(name), maxNameLen),
					map[string]int{"length": len(name), "limit": maxNameLen})
			}

			greeting := "Hello, " + name + "!"
			var warnings []string
			if shout {
				greeting = strings.ToUpper(greeting)
				warnings = append(warnings, "shout mode upper-cased the greeting")
			}
			if g.dryRun {
				warnings = append(warnings, "dry run: nothing was written")
			}

			return &cmdResult{
				data:     map[string]string{"name": name, "greeting": greeting},
				plain:    greeting + "\n",
				warnings: warnings,
			}, nil
		}),
	}

	cmd.Flags().BoolVar(&shout, "shout", false, "upper-case the greeting")
	return cmd
}
