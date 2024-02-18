package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusError = "Error"
)

func Error(msg string) ErrorResponse {
	return ErrorResponse{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) ErrorResponse {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return ErrorResponse{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}

type Credantials struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AccessToken struct {
	Token     string `json:"token" validate:"required"`
	TokenType string `json:"token_type" validate:"required"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Orders []Order

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type OrderSum struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawn struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProccesedAt time.Time `json:"proccesed_at"`
}

type Withdrawals []Withdrawn
