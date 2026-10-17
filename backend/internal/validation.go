package internal

type ValidationCode string

const (
	CodeInvalidID      ValidationCode = "invalid_id"
	CodeInvalidExpense ValidationCode = "invalid_expense"
	CodeInvalidUser    ValidationCode = "invalid_user"
	CodeInvalidFilter  ValidationCode = "invalid_filter"
)

type ValidationError struct {
	Code   ValidationCode
	Field  string
	Reason string
	Cause  error
}

func (e *ValidationError) Error() string {
	message := e.Field + " " + e.Reason
	if e.Cause != nil {
		return message + ": " + e.Cause.Error()
	}
	return message
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}
