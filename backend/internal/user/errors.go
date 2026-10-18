package user

import "errors"

var ErrNotFound = errors.New("user not found")
var ErrTelegramUserIDTaken = errors.New("telegram user ID is already in use")
var ErrHasExpenses = errors.New("user has expenses")
