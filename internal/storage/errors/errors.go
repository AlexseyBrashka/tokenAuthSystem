package errors

import (
	"errors"
)

var (
	NoSuchRefreshToken = errors.New("no such refresh token")
)
