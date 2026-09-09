package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type authorizerStub struct {
	authorize func(token string) (string, error)
}

func (s authorizerStub) Authorize(_ context.Context, token string) (string, error) {
	return s.authorize(token)
}

func TestAuthorize(t *testing.T) {
	auth := authorizerStub{authorize: func(token string) (string, error) {
		if token != "session-token" {
			t.Fatalf("Authorize() token = %q", token)
		}
		return "user-id", nil
	}}
	writeError := func(w http.ResponseWriter, _ *http.Request, _ error) {
		w.WriteHeader(http.StatusUnauthorized)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserID(r.Context()) != "user-id" || Token(r.Context()) != "session-token" {
			t.Errorf("unexpected authorization context")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := Authorize(auth, writeError)(next)

	t.Run("valid bearer token", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
		request.Header.Set("Authorization", "Bearer session-token")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
	})

	t.Run("missing header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		header string
		want   string
		ok     bool
	}{
		{header: "Bearer token", want: "token", ok: true},
		{header: "bearer token", want: "token", ok: true},
		{header: ""},
		{header: "token"},
		{header: "Basic token"},
		{header: "Bearer token extra"},
	}

	for _, test := range tests {
		t.Run(test.header, func(t *testing.T) {
			got, ok := bearerToken(test.header)
			if got != test.want || ok != test.ok {
				t.Errorf("bearerToken(%q) = (%q, %v), want (%q, %v)", test.header, got, ok, test.want, test.ok)
			}
		})
	}
}
