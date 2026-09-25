package cli

import (
	"fmt"
	"runtime"

	"github.com/example/mycli/internal/result"
	"github.com/spf13/cobra"
)

// Build metadata, injected at link time via -ldflags. See the Makefile.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func newVersionCmd(g *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, build date, and runtime",
		Long: "Print build metadata.\n\n" +
			"READS    nothing\n" +
			"WRITES   nothing\n" +
			"NEVER    makes a network call",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if g.json {
				data := map[string]string{
					"version":    Version,
					"commit":     Commit,
					"build_date": BuildDate,
					"go":         runtime.Version(),
				}
				return result.Success("version", data).Write(out)
			}
			_, _ = fmt.Fprintf(out, "mycli %s\ncommit:     %s\nbuild date: %s\ngo:         %s\n",
				Version, Commit, BuildDate, runtime.Version())
			return nil
		},
	}
}
