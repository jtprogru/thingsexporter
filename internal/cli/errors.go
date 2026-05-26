package cli

import "errors"

// ExitCodeError — обёртка над ошибкой, прицепляющая конкретный exit-код.
type ExitCodeError struct {
	Code int
	Err  error
}

// Error реализует error.
func (e *ExitCodeError) Error() string { return e.Err.Error() }

// Unwrap позволяет errors.Is / errors.As добраться до underlying.
func (e *ExitCodeError) Unwrap() error { return e.Err }

// AsExitCode маппит произвольную ошибку в exit-код процесса.
// nil → 0; *ExitCodeError → его Code; иначе → 2.
// Код 1 выдаётся только при панике в main (обрабатывается там через recover).
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

// exit2 — короткий конструктор для ошибок ввода/валидации.
func exit2(err error) *ExitCodeError {
	return &ExitCodeError{Code: 2, Err: err}
}
