package validator

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator(validator *validator.Validate) *CustomValidator {
	return &CustomValidator{validator: validator}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
