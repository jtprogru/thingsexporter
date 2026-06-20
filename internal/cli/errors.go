package cli

import "errors"

// ExitCodeError wraps an error and attaches a specific exit code.
type ExitCodeError struct {
	Code int
	Err  error
}

// Error implements error.
func (e *ExitCodeError) Error() string { return e.Err.Error() }

// Unwrap lets errors.Is / errors.As reach the underlying error.
func (e *ExitCodeError) Unwrap() error { return e.Err }

// AsExitCode maps an arbitrary error to a process exit code.
// nil → 0; *ExitCodeError → its Code; otherwise → 2.
// Code 1 is returned only on a panic in main (handled there via recover).
func AsExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ec *ExitCodeError
	if errors.As(err, &ec) {
		return ec.Code
	}
	return 2
}

// exit2 is a short constructor for input/validation errors.
func exit2(err error) *ExitCodeError {
	return &ExitCodeError{Code: 2, Err: err}
}
