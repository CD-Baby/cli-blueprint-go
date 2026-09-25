package cli

import (
	"errors"

	"github.com/example/mycli/internal/result"
	"github.com/spf13/cobra"
)

// cmdResult is what a command handler returns on success: structured data for
// the --json envelope, plaintext for the default human stdout, and any
// non-fatal warnings to surface in the envelope.
type cmdResult struct {
	data     any
	plain    string
	warnings []string
}

// run wraps a command handler so every command shares one output contract:
//   - success + --json:   one compact result.Success envelope on stdout
//   - success, no --json: the handler's plaintext on stdout
//   - failure + --json:   one compact result.Failure envelope on stdout
//   - failure, any mode:  the *cliError is returned; Execute maps the exit code
//     and cobra prints the diagnostic to stderr (root has SilenceUsage=true).
func run(g *globalFlags, command string,
	fn func(cmd *cobra.Command, args []string) (*cmdResult, error)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		res, err := fn(cmd, args)
		out := cmd.OutOrStdout()
		if err != nil {
			var ce *cliError
			if !errors.As(err, &ce) {
				ce = internalError(err.Error(), nil)
			}
			if g.json {
				_ = result.Failure(command, ce.Result()).Write(out)
			}
			return ce
		}
		var data any
		var warnings []string
		if res != nil {
			data, warnings = res.data, res.warnings
		}
		if g.json {
			_ = result.Success(command, data, warnings...).Write(out)
		} else if res != nil && res.plain != "" {
			_, _ = out.Write([]byte(res.plain))
		}
		return nil
	}
}
