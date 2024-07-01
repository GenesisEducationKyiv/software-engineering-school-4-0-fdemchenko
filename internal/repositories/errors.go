package repositories

import "errors"

var (
	ErrDuplicateEmail = errors.New("email already exists")
)

const PostgreSQLUniqueViolationErrorCode = "23505"
