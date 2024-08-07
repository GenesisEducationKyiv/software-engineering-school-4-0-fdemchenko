package repositories

import "errors"

var (
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrEmailDoesNotExist = errors.New("email does not exist")
)

const PostgreSQLUniqueViolationErrorCode = "23505"
