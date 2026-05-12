package auth

import "errors"

var (
	// ErrPermissionDenied is returned when user lacks required permission
	ErrPermissionDenied = errors.New("permission denied")
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token expired")
)
