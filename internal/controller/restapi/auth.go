package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/request"
	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/response"
	"github.com/maksimovyuriy/astralentry/pkg/formatter"
)

const maxFormBodySize = 1 << 20

func (c *Controller) register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodySize)
	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(formatter.FormatError(http.StatusBadRequest, "invalid request"))
		return
	}

	input := request.Register{
		Token: r.PostForm.Get("token"),
		Login: r.PostForm.Get("login"),
		Pswd:  r.PostForm.Get("pswd"),
	}

	user, err := c.auth.Register(
		r.Context(),
		input.Token,
		input.Login,
		input.Pswd,
	)
	if err != nil {
		c.logger.Error("Register user", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(formatter.FormatError(http.StatusInternalServerError, "internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(formatter.FormatResponse(
		response.Register{Login: user.Login},
	))
}

func (c *Controller) authenticate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodySize)
	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(formatter.FormatError(http.StatusBadRequest, "invalid request"))
		return
	}

	input := request.Auth{
		Login: r.PostForm.Get("login"),
		Pswd:  r.PostForm.Get("pswd"),
	}

	token, err := c.auth.Authenticate(r.Context(), input.Login, input.Pswd)
	if err != nil {
		c.logger.Error("Authenticate user", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(formatter.FormatError(http.StatusInternalServerError, "internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(formatter.FormatResponse(
		response.Auth{Token: token},
	))
}

func (c *Controller) logout(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(formatter.FormatError(http.StatusBadRequest, "invalid request"))
		return
	}

	if err := c.auth.Logout(r.Context(), token); err != nil {
		c.logger.Error("Logout user", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(formatter.FormatError(http.StatusInternalServerError, "internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(formatter.FormatResponse(
		response.Logout{token: true},
	))
}
