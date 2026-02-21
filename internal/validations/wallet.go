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


func ValidateWalletInput(r *http.Request, log zerolog.Logger) (*models.WalletRequest, error) {
	var input_data models.WalletRequest

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
			return nil, formatWalletValidationError(validationErrors)
		}
		return nil, models.ErrInvalidInput
	}

	return &models.WalletRequest{
		TxnId: input_data.TxnId,
		Amount: input_data.Amount,
	}, nil
}

func formatWalletValidationError(errs validator.ValidationErrors) error {
	var errorMessages []string

	for _, err := range errs {
		switch err.Field() {
		case "Amount":
			if err.Tag() == "required" {
				errorMessages = append(errorMessages, "amount is required")
			} else if err.Tag() == "gt" {
				errorMessages = append(errorMessages, "amount must be greater than 0")
			}
		case "TxnId":
			errorMessages = append(errorMessages, "txn_id is required")
		}
	}

	if len(errorMessages) == 0 {
		return models.ErrInvalidInput
	}

	return errors.New(strings.Join(errorMessages, ", "))
}