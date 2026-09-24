package cli

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// logLevels maps the documented --log-level values to slog levels. This table
// is the whole accepted set.
var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// parseLevel resolves a --log-level value. An unknown value is a usage error
// rather than a silent fallback: a typo must not quietly discard the logs the
// caller asked for.
func parseLevel(s string) (slog.Level, error) {
	lvl, ok := logLevels[strings.ToLower(strings.TrimSpace(s))]
	if !ok {
		return 0, usageError(
			fmt.Sprintf("unknown --log-level %q: want debug, info, warn or error", s),
			map[string]string{"value": s})
	}
	return lvl, nil
}

// newLogger builds the logger every command writes diagnostics to. It always
// writes to stderr: stdout carries the result and nothing else.
func newLogger(w io.Writer, lvl slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl}))
}

// log returns the configured logger. It falls back to a discard logger when the
// command tree ran without PersistentPreRunE, so a handler can always log
// without a nil check.
func (g *globalFlags) log() *slog.Logger {
	if g.logger == nil {
		return newLogger(io.Discard, slog.LevelError)
	}
	return g.logger
}
