package domain

import "net/http"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrInvalidRequest   = &AppError{Code: "INVALID_REQUEST", Message: "Invalid request", Status: http.StatusBadRequest}
	ErrUnauthorized     = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", Status: http.StatusUnauthorized}
	ErrForbidden        = &AppError{Code: "FORBIDDEN", Message: "Forbidden", Status: http.StatusForbidden}
	ErrNotFound         = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: http.StatusNotFound}
	ErrConflict         = &AppError{Code: "CONFLICT", Message: "Resource already exists", Status: http.StatusConflict}
	ErrInternal         = &AppError{Code: "INTERNAL_ERROR", Message: "Internal error", Status: http.StatusInternalServerError}
	ErrInvalidSKU       = &AppError{Code: "INVALID_SKU", Message: "Invalid SKU", Status: http.StatusBadRequest}
	ErrDuplicateSKU     = &AppError{Code: "DUPLICATE_SKU", Message: "SKU already exists", Status: http.StatusConflict}
	ErrInvalidLogin     = &AppError{Code: "INVALID_CREDENTIALS", Message: "Invalid username or password", Status: http.StatusUnauthorized}
)
