package http

import (
	"context"
	"errors"
	"events/pkg/events"
	"fmt"
	"net/http"
)

func (a *App) myMiddleware(handler http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("middleware: before handler call")
		handler.ServeHTTP(w, r)
		fmt.Println("middleware: after handler call")
	}

	return http.HandlerFunc(f)
}

func (a *App) addUserToSession(handler http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		exists := a.Session.Exists(r, sessionKeyUser)
		if !exists {
			handler.ServeHTTP(w, r)
			return
		}

		uid := a.Session.GetInt(r, sessionKeyUser)
		if uid == 0 {
			// todo:: handle error
			fmt.Println("error: uid is zero")
			return
		}

		u, err := a.UserService.User(uid)
		if err != nil {
			notFoundErr := &events.ErrNotFound{}
			if errors.As(err, &notFoundErr) {
				a.Session.Remove(r, sessionKeyUser)
				handler.ServeHTTP(w, r)
				return
			}
			a.serverError(w, r, err)
			return
		}

		// creates a new context from the request's existing context
		ctx := context.WithValue(r.Context(), "user", u)

		// manually associate the new context with the request
		r = r.WithContext(ctx)

		// continue call to the handler with the new request
		handler.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (a App) authenticatedUser(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value("user") == nil {
			http.Redirect(w, r, fmt.Sprintf("/login?return_url=%s", r.URL), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}