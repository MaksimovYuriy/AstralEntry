package restapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/response"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
)

var errInvalidRequest = errors.New("invalid request")

func invalidRequest(operation string, cause error) error {
	return fmt.Errorf("%w: %s: %v", errInvalidRequest, operation, cause)
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, text := errorResponse(err)
	logArguments := []any{
		"method", r.Method,
		"route", r.Pattern,
		"status", status,
		"error", err,
	}
	if status >= http.StatusInternalServerError {
		c.logger.Error("HTTP request failed", logArguments...)
	} else {
		c.logger.Warn("HTTP request rejected", logArguments...)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encodeErr := json.NewEncoder(w).Encode(response.FormatError(status, text)); encodeErr != nil {
		c.logger.Error("Encode error response", "error", encodeErr)
	}
}

func errorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, errInvalidRequest):
		return http.StatusBadRequest, errInvalidRequest.Error()

	case errors.Is(err, usecase.ErrInvalidLogin),
		errors.Is(err, usecase.ErrInvalidPassword),
		errors.Is(err, usecase.ErrLoginAlreadyExists),
		errors.Is(err, usecase.ErrGrantUserNotFound),
		errors.Is(err, usecase.ErrUserNotFound),
		errors.Is(err, usecase.ErrInvalidFilter):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, usecase.ErrInvalidAdminToken),
		errors.Is(err, usecase.ErrInvalidCredentials),
		errors.Is(err, usecase.ErrInvalidSession):
		return http.StatusUnauthorized, err.Error()

	case errors.Is(err, usecase.ErrForbidden):
		return http.StatusForbidden, err.Error()

	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
