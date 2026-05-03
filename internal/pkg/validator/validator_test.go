package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0,lte=130"`
}

func TestValidate(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		valid := TestStruct{
			Name:  "John",
			Email: "john@example.com",
			Age:   25,
		}
		err := Validate(valid)
		assert.Nil(t, err)
	})

	t.Run("missing required field", func(t *testing.T) {
		missingName := TestStruct{
			Email: "john@example.com",
			Age:   25,
		}
		err := Validate(missingName)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Name")
	})

	t.Run("invalid email format", func(t *testing.T) {
		invalidEmail := TestStruct{
			Name:  "John",
			Email: "not-an-email",
			Age:   25,
		}
		err := Validate(invalidEmail)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Email")
	})

	t.Run("age out of range - negative", func(t *testing.T) {
		negativeAge := TestStruct{
			Name:  "John",
			Email: "john@example.com",
			Age:   -1,
		}
		err := Validate(negativeAge)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Age")
	})

	t.Run("age out of range - too high", func(t *testing.T) {
		highAge := TestStruct{
			Name:  "John",
			Email: "john@example.com",
			Age:   200,
		}
		err := Validate(highAge)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Age")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		multipleErrors := TestStruct{
			Age: 200,
		}
		err := Validate(multipleErrors)
		assert.NotNil(t, err)
	})
}

func TestValidateVar(t *testing.T) {
	t.Run("valid variable", func(t *testing.T) {
		err := ValidateVar("test@example.com", "required,email")
		assert.Nil(t, err)
	})

	t.Run("invalid email variable", func(t *testing.T) {
		err := ValidateVar("not-an-email", "email")
		assert.NotNil(t, err)
	})

	t.Run("required empty string", func(t *testing.T) {
		err := ValidateVar("", "required")
		assert.NotNil(t, err)
	})

	t.Run("valid url", func(t *testing.T) {
		err := ValidateVar("https://example.com/path", "url")
		assert.Nil(t, err)
	})

	t.Run("invalid url", func(t *testing.T) {
		err := ValidateVar("not-a-url", "url")
		assert.NotNil(t, err)
	})

	t.Run("min length", func(t *testing.T) {
		err := ValidateVar("ab", "min=3")
		assert.NotNil(t, err)

		err = ValidateVar("abc", "min=3")
		assert.Nil(t, err)
	})

	t.Run("max length", func(t *testing.T) {
		err := ValidateVar("abcdefghijk", "max=10")
		assert.NotNil(t, err)

		err = ValidateVar("abc", "max=10")
		assert.Nil(t, err)
	})
}
