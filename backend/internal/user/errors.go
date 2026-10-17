package user

import "errors"

var ErrNotFound = errors.New("user not found")
var ErrTelegramUserIDTaken = errors.New("telegram user id already belongs to another user")
var ErrHasExpenses = errors.New("user has expenses")
