package validations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/AdityaTote/wallet-service/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

func ValidateAuthInput(r *http.Request, log zerolog.Logger) (*models.UserParams, error) {
	var input_data models.UserParams

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&input_data); err != nil {
		return nil, models.ErrInvalidBody
	}

	validate := validator.New()

	err := validate.Struct(input_data)
	if err != nil {
		log.Error().Err(err).Msg("validation failed for auth input validation")

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return nil, formatValidationError(validationErrors)
		}
		return nil, models.ErrInvalidInput
	}

	return &models.UserParams{
		Username:    input_data.Username,
		Password: input_data.Password,
	}, nil
}

func formatValidationError(errs validator.ValidationErrors) error {
	var errorMessages []string

	for _, err := range errs {
		switch err.Field() {
		case "Username":
				errorMessages = append(errorMessages, "email is required")
		case "Password":
			if err.Tag() == "required" {
				errorMessages = append(errorMessages, "password is required")
			} else if err.Tag() == "min" {
				errorMessages = append(errorMessages, "password must be at least 8 characters long")
			}
		}
	}
		if len(errorMessages) == 0 {
		return models.ErrInvalidInput
	}

	return errors.New(strings.Join(errorMessages, ", "))

}