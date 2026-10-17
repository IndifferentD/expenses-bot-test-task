package internal

import "errors"

var ErrExpenseNotFound = errors.New("expense not found")
var ErrUserNotFound = errors.New("user not found")
var ErrTelegramUserIDTaken = errors.New("telegram user id already belongs to another user")
var ErrUserHasExpenses = errors.New("user has expenses")

type ValidationCode string

const (
	CodeInvalidID      ValidationCode = "invalid_id"
	CodeInvalidExpense ValidationCode = "invalid_expense"
	CodeInvalidUser    ValidationCode = "invalid_user"
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
