package cli

import "github.com/kpearson/mycli/internal/result"

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

func internalError(msg string, detail any) *cliError {
	return newErr(ExitInternal, "INTERNAL_ERROR", msg, detail)
}
