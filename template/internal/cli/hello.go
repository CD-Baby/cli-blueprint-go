package cli

import (
	"errors"
	"os"

	"github.com/example/mycli/internal/greeting"
	"github.com/spf13/cobra"
)

// newHelloCmd is the worked example of the command pattern. Copy it when you
// add a command, then delete it. It shows all six pieces:
//
//  1. run(g, "<name>", ...) wraps the handler in the shared output contract
//  2. the handler holds no logic: internal/greeting does the work
//  3. the handler returns a *cmdResult, and never writes to stdout itself
//  4. domain errors are mapped onto the exit code contract here, not there
//  5. cmdResult.data feeds --json; cmdResult.plain feeds human stdout
//  6. warnings ride along in the envelope without failing the command
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
			g.log().Debug("resolved name", "name", name, "from_env", len(args) == 0)

			text, err := greeting.Compose(name, greeting.Options{Shout: shout})
			if err != nil {
				return nil, helloFailure(err)
			}

			var warnings []string
			if shout {
				warnings = append(warnings, "shout mode upper-cased the greeting")
			}
			if g.dryRun {
				g.log().Warn("dry run: skipping writes")
				warnings = append(warnings, "dry run: nothing was written")
			}

			return &cmdResult{
				data:     map[string]string{"name": name, "greeting": text},
				plain:    text + "\n",
				warnings: warnings,
			}, nil
		}),
	}

	cmd.Flags().BoolVar(&shout, "shout", false, "upper-case the greeting")
	return cmd
}

// helloFailure translates a greeting error into the exit code contract. This
// mapping belongs to the command layer: the domain package decides what is
// wrong, and the command decides what that costs the caller.
func helloFailure(err error) error {
	var tooLong *greeting.ErrNameTooLong
	switch {
	case errors.Is(err, greeting.ErrNoName):
		return prereqError(
			"no name given: pass one as an argument or set MYCLI_NAME", nil)
	case errors.As(err, &tooLong):
		return validationError(tooLong.Error(),
			map[string]int{"length": tooLong.Len, "limit": tooLong.Limit})
	default:
		return internalError(err.Error(), nil)
	}
}
