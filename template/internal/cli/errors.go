package cli

import "github.com/example/mycli/internal/result"

// cliError carries both the process exit code a command should terminate with
// and the envelope payload printed under --json. RunE handlers return these;
// Execute maps the exit code (cliError satisfies CodedError).
type cliError struct {
	code int
	rerr *result.Error
}

func (e *cliError) Error() string         { return e.rerr.Message }
func (e *cliError) ExitCode() int         { return e.code }
func (e *cliError) Result() *result.Error { return e.rerr }

func newErr(code int, errCode, msg string, detail any) *cliError {
	return &cliError{code: code, rerr: &result.Error{Code: errCode, Message: msg, Detail: detail}}
}

func usageError(msg string, detail any) *cliError {
	return newErr(ExitUsage, "USAGE_ERROR", msg, detail)
}

func validationError(msg string, detail any) *cliError {
	return newErr(ExitValidation, "VALIDATION_FAILED", msg, detail)
}

func prereqError(msg string, detail any) *cliError {
	return newErr(ExitPrereq, "MISSING_PREREQUISITE", msg, detail)
}

// canceledError reports that a signal stopped the command. run() produces it
// from a context error, so handlers only have to return ctx.Err().
func canceledError() *cliError {
	return newErr(ExitCanceled, "CANCELED", "canceled by signal", nil)
}

func internalError(msg string, detail any) *cliError {
	return newErr(ExitInternal, "INTERNAL_ERROR", msg, detail)
}
