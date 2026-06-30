package monime

import "github.com/go-playground/validator/v10"

// validate is the shared validator instance used across all resources.
var validate = validator.New(validator.WithRequiredStructEnabled())

// validateStruct runs struct-tag validation on s and converts any failure into
// a *ValidationError.
func validateStruct(s any) error {
	if err := validate.Struct(s); err != nil {
		return newValidationError(err.Error())
	}
	return nil
}
