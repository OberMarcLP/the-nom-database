package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/models"
)

// newRequestWithVars builds a test request and attaches gorilla/mux path
// variables so handlers can read them via mux.Vars.
func newRequestWithVars(method, target string, body io.Reader, vars map[string]string) *http.Request {
	r := httptest.NewRequest(method, target, body)
	if vars != nil {
		r = mux.SetURLVars(r, vars)
	}
	return r
}

// newJSONRequestWithVars is like newRequestWithVars but sends a JSON body.
func newJSONRequestWithVars(method, target, body string, vars map[string]string) *http.Request {
	r := newRequestWithVars(method, target, strings.NewReader(body), vars)
	r.Header.Set("Content-Type", "application/json")
	return r
}

// withUser returns a copy of the request carrying the given user in the
// context, mirroring what the auth middleware does.
func withUser(r *http.Request, user *models.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), models.UserContextKey, user))
}
