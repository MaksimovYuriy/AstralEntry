package restapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/response"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
)

var errInvalidRequest = errors.New("invalid request")

func (c *Controller) writeError(w http.ResponseWriter, err error) {
	status, text := errorResponse(err)
	if status == http.StatusInternalServerError {
		c.logger.Error("Handle HTTP request", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encodeErr := json.NewEncoder(w).Encode(response.FormatError(status, text)); encodeErr != nil {
		c.logger.Error("Encode error response", "error", encodeErr)
	}
}

func errorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, errInvalidRequest),
		errors.Is(err, usecase.ErrInvalidLogin),
		errors.Is(err, usecase.ErrInvalidPassword),
		errors.Is(err, usecase.ErrLoginAlreadyExists),
		errors.Is(err, usecase.ErrGrantUserNotFound):
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
