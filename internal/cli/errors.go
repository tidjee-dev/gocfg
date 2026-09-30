package cli

// ExitError carries a process exit code. Validation/config failures use 1,
// CLI usage errors use 2 (see ROADMAP §28).
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }
