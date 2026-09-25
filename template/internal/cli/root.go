// Package cli defines the mycli command tree, global flags, and the exit code
// contract. main() calls Execute.
package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Process exit codes. This table is the contract documented in root --help;
// callers branch on these values, and the code always agrees with the
// envelope's "ok" field.
const (
	ExitOK         = 0 // success
	ExitInternal   = 1 // bug or I/O failure
	ExitUsage      = 2 // unknown command/flag, bad argument
	ExitValidation = 3 // input failed schema or semantic checks
	ExitPrereq     = 4 // missing prerequisite (no input, unknown target, ...)

	// ExitCanceled reports that a signal stopped the command. It follows the
	// shell convention of 128 + the signal number, and 2 is SIGINT.
	ExitCanceled = 130
)

// CodedError is an error that carries the process exit code a command should
// terminate with. RunE handlers return these; Execute maps them to os.Exit.
type CodedError interface {
	error
	ExitCode() int
}

// globalFlags holds the values of the persistent flags bound on the root
// command. Commands read it to honor --json, --dry-run, etc.
type globalFlags struct {
	json     bool
	logLevel string
	quiet    bool
	dryRun   bool

	// logger writes diagnostics to stderr. PersistentPreRunE builds it once the
	// level is resolved; read it through g.log(), never directly.
	logger *slog.Logger
}

// Execute builds and runs the root command, returning the process exit code.
//
// SIGINT and SIGTERM cancel the command's context instead of killing the
// process, so in-flight work unwinds and deferred cleanup runs. A second
// signal is left to the default handler, which terminates immediately: a
// command that ignores the first signal must not become unkillable.
func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return execute(ctx, newRootCmd())
}

// execute runs an already-built command tree and maps its error to an exit
// code. Kept separate from Execute so tests can drive a root with its args,
// output streams and context under their own control.
func execute(ctx context.Context, root *cobra.Command) int {
	err := root.ExecuteContext(ctx)
	if err == nil {
		return ExitOK
	}
	var ce CodedError
	if errors.As(err, &ce) {
		return ce.ExitCode()
	}
	// Flag and argument parse errors reach here; cobra has already reported them.
	return ExitUsage
}

func newRootCmd() *cobra.Command {
	return newRootCmdWith(&globalFlags{})
}

// newRootCmdWith builds the command tree around a caller-supplied globalFlags,
// so tests can inspect what the persistent flags resolved to.
func newRootCmdWith(g *globalFlags) *cobra.Command {
	root := &cobra.Command{
		Use:   "mycli",
		Short: "One-line description of what this CLI does",
		Long: "Longer description.\n\n" +
			"EXIT CODES\n" +
			"    0  success\n" +
			"    1  internal error (bug or I/O failure)\n" +
			"    2  usage error (unknown command/flag, bad argument)\n" +
			"    3  validation failure (input failed schema or semantic checks)\n" +
			"    4  missing prerequisite\n" +
			"  130  canceled by SIGINT or SIGTERM",
		SilenceUsage:  true,
		SilenceErrors: false,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			applyEnv(cmd.Flags())
			if g.quiet {
				g.logLevel = "error"
			}
			lvl, err := parseLevel(g.logLevel)
			if err != nil {
				return err
			}
			g.logger = newLogger(cmd.ErrOrStderr(), lvl)
			return nil
		},
	}

	pf := root.PersistentFlags()
	pf.BoolVar(&g.json, "json", false, "emit a machine-readable result envelope on stdout")
	pf.StringVar(&g.logLevel, "log-level", "info", "stderr log verbosity: debug, info, warn, error")
	pf.BoolVar(&g.quiet, "quiet", false, "equivalent to --log-level error")
	pf.BoolVar(&g.dryRun, "dry-run", false, "where supported: full validation, no writes")

	root.AddCommand(newVersionCmd(g))
	root.AddCommand(newHelloCmd(g))
	return root
}

// envBindings maps a persistent flag to the environment variable that supplies
// its default. Add a row here when you add an env-configurable flag.
var envBindings = []struct{ flag, env string }{
	{"log-level", "MYCLI_LOG"},
}

// applyEnv fills flags from the environment, preserving flag > env > default
// precedence: a flag the user set explicitly is never overwritten.
func applyEnv(fs *pflag.FlagSet) {
	v := viper.New()
	for _, b := range envBindings {
		_ = v.BindEnv(b.flag, b.env)
		if fs.Changed(b.flag) {
			continue
		}
		if val := v.GetString(b.flag); val != "" {
			_ = fs.Set(b.flag, val)
		}
	}
}
