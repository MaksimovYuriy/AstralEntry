package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/request"
	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/response"
)

const maxFormBodySize = 1 << 20

func (c *Controller) register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodySize)
	if err := r.ParseForm(); err != nil {
		c.writeError(w, errInvalidRequest)
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
		c.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response.FormatResponse(
		response.Register{Login: user.Login},
	))
}

func (c *Controller) authenticate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodySize)
	if err := r.ParseForm(); err != nil {
		c.writeError(w, errInvalidRequest)
		return
	}

	input := request.Auth{
		Login: r.PostForm.Get("login"),
		Pswd:  r.PostForm.Get("pswd"),
	}

	token, err := c.auth.Authenticate(r.Context(), input.Login, input.Pswd)
	if err != nil {
		c.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response.FormatResponse(
		response.Auth{Token: token},
	))
}

func (c *Controller) logout(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		c.writeError(w, errInvalidRequest)
		return
	}

	if err := c.auth.Logout(r.Context(), token); err != nil {
		c.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response.FormatResponse(
		response.Logout{token: true},
	))
}
