package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("Error without underlying error", func(t *testing.T) {
		appErr := &AppError{
			Code:    "NOT_FOUND",
			Message: "resource not found",
		}
		assert.Equal(t, "resource not found", appErr.Error())
	})

	t.Run("Error with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("database connection failed")
		appErr := &AppError{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch resource",
			Err:     underlyingErr,
		}
		assert.Equal(t, "failed to fetch resource: database connection failed", appErr.Error())
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		appErr := &AppError{
			Code:    "TEST",
			Message: "test message",
			Err:     underlyingErr,
		}
		assert.Equal(t, underlyingErr, appErr.Unwrap())
	})

	t.Run("Unwrap returns nil when no underlying error", func(t *testing.T) {
		appErr := &AppError{
			Code:    "TEST",
			Message: "test message",
		}
		assert.Nil(t, appErr.Unwrap())
	})
}

func TestNewAppError(t *testing.T) {
	underlyingErr := errors.New("base error")
	appErr := NewAppError("CODE_001", "something went wrong", underlyingErr)

	assert.Equal(t, "CODE_001", appErr.Code)
	assert.Equal(t, "something went wrong", appErr.Message)
	assert.Equal(t, underlyingErr, appErr.Err)
}

func TestSentinelErrors(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		assert.Equal(t, "resource not found", ErrNotFound.Error())
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {
		assert.Equal(t, "unauthorized", ErrUnauthorized.Error())
	})

	t.Run("ErrBadRequest", func(t *testing.T) {
		assert.Equal(t, "bad request", ErrBadRequest.Error())
	})

	t.Run("ErrInternalServer", func(t *testing.T) {
		assert.Equal(t, "internal server error", ErrInternalServer.Error())
	})
}

func TestErrorsIs(t *testing.T) {
	t.Run("errors.Is works with sentinel errors", func(t *testing.T) {
		err := ErrNotFound
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.False(t, errors.Is(err, ErrUnauthorized))
	})

	t.Run("errors.Is works with wrapped AppError", func(t *testing.T) {
		appErr := NewAppError("NOT_FOUND", "user not found", ErrNotFound)
		assert.True(t, errors.Is(appErr, ErrNotFound))
	})
}
